#!/bin/bash

# Choovio IoT Platform - EC2 User Data Script
# This script automatically sets up the platform on new EC2 instances

set -e

# Variables from Terraform
DB_HOST="${db_host}"
DB_NAME="${db_name}"
DB_USERNAME="${db_username}"
DB_PASSWORD="${db_password}"
AWS_REGION="${aws_region}"

# Logging
exec > >(tee /var/log/user-data.log|logger -t user-data -s 2>/dev/console) 2>&1
echo "Starting Choovio IoT Platform deployment at $(date)"

# Update system
yum update -y

# Install required packages
yum install -y \
    docker \
    git \
    curl \
    wget \
    unzip \
    htop \
    nginx

# Install Docker Compose
curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
chmod +x /usr/local/bin/docker-compose

# Start and enable Docker
systemctl start docker
systemctl enable docker
usermod -a -G docker ec2-user

# Install Node.js 18
curl -fsSL https://rpm.nodesource.com/setup_18.x | bash -
yum install -y nodejs

# Install Go 1.21
wget https://golang.org/dl/go1.21.0.linux-amd64.tar.gz
tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/profile
export PATH=$PATH:/usr/local/go/bin

# Create application directory
mkdir -p /opt/choovio
cd /opt/choovio

# Clone the repository (replace with your actual repository URL)
git clone https://github.com/isrivarshini/Choovio-Project.git .
cd /opt/choovio

# Set up environment variables
cat > /opt/choovio/.env << EOF
# Database Configuration
DB_HOST=${DB_HOST}
DB_PORT=5432
DB_NAME=${DB_NAME}
DB_USERNAME=${DB_USERNAME}
DB_PASSWORD=${DB_PASSWORD}

# Application Configuration
APP_ENV=production
AWS_REGION=${AWS_REGION}

# Magistrala Configuration
MG_DB_HOST=${DB_HOST}
MG_DB_PORT=5432
MG_DB_USER=${DB_USERNAME}
MG_DB_PASS=${DB_PASSWORD}
MG_DB_NAME=${DB_NAME}

# API Configuration
API_BASE_URL=http://localhost:9000
FRONTEND_URL=http://localhost:5173

# CORS Configuration
CORS_ALLOWED_ORIGINS=*
EOF

# Build and start backend services
echo "Setting up backend services..."
cd /opt/choovio/backend

# Build the backend binary
/usr/local/go/bin/go mod tidy
/usr/local/go/bin/go build -o magistrala-api ./cmd/main.go

# Create systemd service for backend
cat > /etc/systemd/system/choovio-backend.service << EOF
[Unit]
Description=Choovio IoT Backend API
After=network.target

[Service]
Type=simple
User=ec2-user
WorkingDirectory=/opt/choovio/backend
Environment=PATH=/usr/local/go/bin:/usr/bin:/bin
EnvironmentFile=/opt/choovio/.env
ExecStart=/opt/choovio/Backend/magistrala-api
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
EOF

# Set up frontend
echo "Setting up frontend..."
cd /opt/choovio/Frontend

# Install dependencies and build
npm install
npm run build

# Configure Nginx for frontend
cat > /etc/nginx/conf.d/choovio.conf << EOF
server {
    listen 5173;
    server_name _;
    root /opt/choovio/Frontend/dist;
    index index.html;

    location / {
        try_files \$uri \$uri/ /index.html;
    }

    location /api/ {
        proxy_pass http://localhost:9000/;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
    }

    location /health {
        proxy_pass http://localhost:9000/health;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
    }
}
EOF

# Remove default Nginx configuration
rm -f /etc/nginx/conf.d/default.conf

# Set up Docker Compose for Magistrala core services
cd /opt/choovio
cat > docker-compose.override.yml << EOF
version: '3.8'

services:
  postgres:
    environment:
      POSTGRES_HOST: ${DB_HOST}
      POSTGRES_PORT: 5432
      POSTGRES_USER: ${DB_USERNAME}
      POSTGRES_PASSWORD: ${DB_PASSWORD}
      POSTGRES_DB: ${DB_NAME}
    
  redis:
    restart: always
    
  nats:
    restart: always

  things:
    environment:
      MG_THINGS_DB_HOST: ${DB_HOST}
      MG_THINGS_DB_PORT: 5432
      MG_THINGS_DB_USER: ${DB_USERNAME}
      MG_THINGS_DB_PASS: ${DB_PASSWORD}
      MG_THINGS_DB_NAME: ${DB_NAME}
    
  users:
    environment:
      MG_USERS_DB_HOST: ${DB_HOST}
      MG_USERS_DB_PORT: 5432
      MG_USERS_DB_USER: ${DB_USERNAME}
      MG_USERS_DB_PASS: ${DB_PASSWORD}
      MG_USERS_DB_NAME: ${DB_NAME}
EOF

# Wait for database to be ready
echo "Waiting for database to be ready..."
while ! nc -z ${DB_HOST} 5432; do
  echo "Waiting for PostgreSQL to start..."
  sleep 5
done

# Start core Magistrala services
cd /opt/choovio/docker
docker-compose up -d postgres redis nats

# Start Magistrala services
sleep 30  # Wait for dependencies
docker-compose up -d users things http-adapter

# Set permissions
chown -R ec2-user:ec2-user /opt/choovio

# Start services
systemctl daemon-reload
systemctl enable choovio-backend
systemctl start choovio-backend

systemctl enable nginx
systemctl start nginx

# Create health check script
cat > /opt/choovio/health-check.sh << 'EOF'
#!/bin/bash
# Health check script for monitoring services

echo "=== Choovio IoT Platform Health Check ==="
echo "Timestamp: $(date)"

echo "1. Backend API Status:"
curl -s http://localhost:9000/health || echo "Backend API is DOWN"

echo "2. Frontend Status:"
curl -s http://localhost:5173 > /dev/null && echo "Frontend is UP" || echo "Frontend is DOWN"

echo "3. Docker Services:"
docker-compose -f /opt/choovio/Docker/docker-compose.yml ps

echo "4. System Resources:"
free -h
df -h /

echo "=== End Health Check ==="
EOF

chmod +x /opt/choovio/health-check.sh

# Set up log rotation
cat > /etc/logrotate.d/choovio << EOF
/var/log/choovio/*.log {
    daily
    missingok
    rotate 14
    compress
    notifempty
    create 0644 ec2-user ec2-user
    postrotate
        systemctl reload choovio-backend
    endscript
}
EOF

# Create CloudWatch agent configuration (optional)
cat > /opt/aws/amazon-cloudwatch-agent/etc/amazon-cloudwatch-agent.json << EOF
{
    "logs": {
        "logs_collected": {
            "files": {
                "collect_list": [
                    {
                        "file_path": "/var/log/user-data.log",
                        "log_group_name": "choovio-iot/user-data",
                        "log_stream_name": "{instance_id}"
                    },
                    {
                        "file_path": "/var/log/choovio/*.log",
                        "log_group_name": "choovio-iot/application",
                        "log_stream_name": "{instance_id}"
                    }
                ]
            }
        }
    }
}
EOF

# Final status check
echo "Deployment completed at $(date)"
echo "Running final health check..."
sleep 60  # Wait for services to fully start
/opt/choovio/health-check.sh

echo "Choovio IoT Platform deployment completed successfully!"
echo "Frontend URL: http://$(curl -s http://169.254.169.254/latest/meta-data/public-ipv4):5173"
echo "Backend API: http://$(curl -s http://169.254.169.254/latest/meta-data/public-ipv4):9000" 