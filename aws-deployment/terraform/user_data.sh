#!/bin/bash

# Choovio IoT Platform - FREE TIER EC2 User Data Script
# Optimized for minimal cost deployment

set -e

# Variables from Terraform
PROJECT_NAME="${project_name}"
AWS_REGION="${aws_region}"

# Logging
exec > >(tee /var/log/user-data.log|logger -t user-data -s 2>/dev/console) 2>&1
echo "Starting Choovio IoT Platform FREE TIER deployment at $(date)"

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
    nginx \
    sqlite \
    postgresql \
    postgresql-server

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

# Clone the repository
git clone https://github.com/isrivarshini/Choovio-Project.git .
cd /opt/choovio

# Set up local PostgreSQL (FREE - no RDS charges)
postgresql-setup initdb
systemctl enable postgresql
systemctl start postgresql

# Create database and user
sudo -u postgres psql << EOF
CREATE DATABASE magistrala;
CREATE USER magistrala WITH PASSWORD 'magistrala123!';
GRANT ALL PRIVILEGES ON DATABASE magistrala TO magistrala;
\q
EOF

# Set up environment variables for local deployment
cat > /opt/choovio/.env << EOF
# Local Database Configuration (FREE)
DB_HOST=localhost
DB_PORT=5432
DB_NAME=magistrala
DB_USERNAME=magistrala
DB_PASSWORD=magistrala123!

# Application Configuration
APP_ENV=development
AWS_REGION=${aws_region}
PROJECT_NAME=${project_name}

# Magistrala Configuration - Local
MG_DB_HOST=localhost
MG_DB_PORT=5432
MG_DB_USER=magistrala
MG_DB_PASS=magistrala123!
MG_DB_NAME=magistrala

# API Configuration
API_BASE_URL=http://localhost:9000
FRONTEND_URL=http://localhost:80

# CORS Configuration
CORS_ALLOWED_ORIGINS=*

# Free Tier Optimizations
DEPLOYMENT_TYPE=free_tier
USE_LOCAL_DB=true
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
After=network.target postgresql.service

[Service]
Type=simple
User=ec2-user
WorkingDirectory=/opt/choovio/backend
Environment=PATH=/usr/local/go/bin:/usr/bin:/bin
EnvironmentFile=/opt/choovio/.env
ExecStart=/opt/choovio/backend/magistrala-api
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
EOF

# Set up frontend
echo "Setting up frontend..."
cd /opt/choovio/Frontend

# Update frontend API configuration for production
cat > /opt/choovio/Frontend/.env.production << EOF
VITE_API_BASE_URL=http://\$(curl -s http://169.254.169.254/latest/meta-data/public-ipv4):9000
VITE_DEPLOYMENT_TYPE=free_tier
EOF

# Install dependencies and build
npm install
npm run build

# Configure Nginx for frontend
cat > /etc/nginx/nginx.conf << EOF
user nginx;
worker_processes auto;
error_log /var/log/nginx/error.log;
pid /run/nginx.pid;

events {
    worker_connections 1024;
}

http {
    log_format  main  '\$remote_addr - \$remote_user [\$time_local] "\$request" '
                      '\$status \$body_bytes_sent "\$http_referer" '
                      '"\$http_user_agent" "\$http_x_forwarded_for"';

    access_log  /var/log/nginx/access.log  main;

    sendfile            on;
    tcp_nopush          on;
    tcp_nodelay         on;
    keepalive_timeout   65;
    types_hash_max_size 2048;

    include             /etc/nginx/mime.types;
    default_type        application/octet-stream;

    # Frontend server
    server {
        listen 80 default_server;
        server_name _;
        root /opt/choovio/Frontend/dist;
        index index.html;

        # Frontend routing
        location / {
            try_files \$uri \$uri/ /index.html;
        }

        # API proxy
        location /api/ {
            proxy_pass http://localhost:9000/;
            proxy_set_header Host \$host;
            proxy_set_header X-Real-IP \$remote_addr;
            proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto \$scheme;
        }

        # Direct backend access
        location /health {
            proxy_pass http://localhost:9000/health;
            proxy_set_header Host \$host;
            proxy_set_header X-Real-IP \$remote_addr;
        }

        # Static assets
        location /assets/ {
            expires 1y;
            add_header Cache-Control "public, immutable";
        }
    }

    # Backend direct access (for development)
    server {
        listen 9000;
        server_name _;
        
        location / {
            proxy_pass http://localhost:9001;
            proxy_set_header Host \$host;
            proxy_set_header X-Real-IP \$remote_addr;
        }
    }
}
EOF

# Set up minimal Docker services (only essential ones for free tier)
cd /opt/choovio
cat > docker-compose.free-tier.yml << EOF
version: '3.8'

services:
  # Redis for caching (lightweight)
  redis:
    image: redis:7-alpine
    restart: unless-stopped
    ports:
      - "6379:6379"
    command: redis-server --appendonly yes
    volumes:
      - redis_data:/data

  # NATS for messaging (lightweight)
  nats:
    image: nats:2-alpine
    restart: unless-stopped
    ports:
      - "4222:4222"
    command: 
      - "-js"
      - "-m"
      - "8222"

volumes:
  redis_data:
EOF

# Wait for PostgreSQL to be ready
echo "Waiting for PostgreSQL to be ready..."
while ! nc -z localhost 5432; do
  echo "Waiting for PostgreSQL to start..."
  sleep 5
done

# Start minimal Docker services
docker-compose -f docker-compose.free-tier.yml up -d

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
# Health check script for FREE TIER deployment

echo "=== Choovio IoT Platform Health Check (Free Tier) ==="
echo "Timestamp: $(date)"

echo "1. Frontend Status:"
curl -s http://localhost > /dev/null && echo "✅ Frontend is UP" || echo "❌ Frontend is DOWN"

echo "2. Backend API Status:"
curl -s http://localhost:9000/health > /dev/null && echo "✅ Backend API is UP" || echo "❌ Backend API is DOWN"

echo "3. Database Status:"
sudo -u postgres psql -d magistrala -c "SELECT 1;" > /dev/null 2>&1 && echo "✅ Database is UP" || echo "❌ Database is DOWN"

echo "4. Docker Services:"
docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"

echo "5. System Resources:"
echo "Memory: $(free -h | grep Mem | awk '{print $3"/"$2}')"
echo "Disk: $(df -h / | tail -1 | awk '{print $3"/"$2" ("$5" used)"}')"
echo "CPU Load: $(uptime | awk -F'load average:' '{print $2}')"

echo "6. Public Access:"
PUBLIC_IP=$(curl -s http://169.254.169.254/latest/meta-data/public-ipv4)
echo "🌐 Frontend URL: http://$PUBLIC_IP"
echo "🔧 Backend API: http://$PUBLIC_IP:9000"

echo "=== End Health Check ==="
EOF

chmod +x /opt/choovio/health-check.sh

# Create startup script for easy management
cat > /opt/choovio/manage.sh << 'EOF'
#!/bin/bash
# Management script for Choovio IoT Platform

case "$1" in
    start)
        echo "Starting Choovio services..."
        systemctl start postgresql choovio-backend nginx
        docker-compose -f /opt/choovio/docker-compose.free-tier.yml up -d
        echo "Services started!"
        ;;
    stop)
        echo "Stopping Choovio services..."
        systemctl stop choovio-backend nginx
        docker-compose -f /opt/choovio/docker-compose.free-tier.yml down
        echo "Services stopped!"
        ;;
    restart)
        $0 stop
        sleep 5
        $0 start
        ;;
    status)
        /opt/choovio/health-check.sh
        ;;
    logs)
        echo "=== Backend Logs ==="
        journalctl -u choovio-backend -n 50
        echo "=== Nginx Logs ==="
        tail -n 20 /var/log/nginx/error.log
        ;;
    *)
        echo "Usage: $0 {start|stop|restart|status|logs}"
        exit 1
        ;;
esac
EOF

chmod +x /opt/choovio/manage.sh

# Final setup and verification
echo "Performing final setup..."
sleep 30  # Wait for services to stabilize

# Update frontend build with correct API URL
PUBLIC_IP=$(curl -s http://169.254.169.254/latest/meta-data/public-ipv4)
echo "VITE_API_BASE_URL=http://$PUBLIC_IP:9000" > /opt/choovio/Frontend/.env.production

# Rebuild frontend with correct API URL
cd /opt/choovio/Frontend
npm run build

# Restart nginx to serve updated frontend
systemctl restart nginx

# Final health check
echo "Deployment completed at $(date)"
echo "Running final health check..."
sleep 30
/opt/choovio/health-check.sh

echo ""
echo "🎉 Choovio IoT Platform FREE TIER deployment completed successfully!"
echo "📱 Frontend URL: http://$PUBLIC_IP"
echo "🔧 Backend API: http://$PUBLIC_IP:9000"
echo "💰 Estimated monthly cost: \$0-5 (within free tier)"
echo ""
echo "🔍 Management commands:"
echo "  sudo /opt/choovio/manage.sh status  - Check system status"
echo "  sudo /opt/choovio/manage.sh restart - Restart all services"
echo "  sudo /opt/choovio/manage.sh logs    - View application logs"
echo ""
echo "🔑 Default login credentials:"
echo "  Email: admin@example.com"
echo "  Password: admin123"
EOF 