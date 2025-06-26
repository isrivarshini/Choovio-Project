# 🚀 Choovio IoT Platform

A modern, enterprise-grade IoT platform built on Magistrala with custom React frontend and AWS deployment capabilities.

![Choovio Logo](Frontend/public/assets/logo1.webp)

## 📋 Project Overview

The Choovio IoT Platform is a comprehensive pilot project demonstrating:
- **Modern IoT Infrastructure**: Built on the robust Magistrala platform
- **Custom Branding**: White-label solution with Choovio branding
- **Enterprise Frontend**: React + TypeScript admin dashboard
- **Cloud-Ready**: Production AWS deployment with auto-scaling
- **AI-Assisted Development**: Extensively documented AI usage throughout

---

## 🏗️ Project Structure

```
magistrala/
├── Backend/              # Magistrala IoT platform (Go)
├── Frontend/             # React admin dashboard (TypeScript + Vite)
├── Docker/               # Local development containers
├── aws-deployment/       # AWS infrastructure (Terraform)
│   ├── terraform/        # Infrastructure as Code
│   ├── scripts/          # Deployment automation
│   └── docs/             # AWS deployment guides
├── PROJECT_REPORT.md     # Comprehensive project documentation
└── README.md             # This file
```

---

## ⚡ Quick Start (Local Development)

### Prerequisites
- Docker & Docker Compose
- Node.js 18+
- npm or yarn

### 1. Start Backend Services
```bash
cd Docker
docker-compose up -d
```

### 2. Start Frontend Dashboard
```bash
cd Frontend
npm install
npm run dev
```

### 3. Access the Application
- **Dashboard**: http://localhost:5173
- **Backend API**: http://localhost:9000

### Demo Credentials
- **Email**: admin@example.com
- **Password**: 12345678
- **Role**: Admin (Full access)

---

## 🌐 AWS Deployment (Production)

Deploy your Choovio IoT Platform to AWS and make it publicly accessible.

### Prerequisites for AWS Deployment

1. **AWS Account**: Create at https://aws.amazon.com/
2. **AWS CLI**: Install and configure with your credentials
3. **Terraform**: Version 1.0+ installed
4. **IAM Permissions**: User with EC2, RDS, VPC, ELB, Auto Scaling access

### Step-by-Step Deployment

#### 1. Configure AWS Credentials
```bash
aws configure
```
Enter your:
- AWS Access Key ID
- AWS Secret Access Key
- Default region: `us-east-1`
- Default output format: `json`

#### 2. Verify AWS Setup
```bash
aws sts get-caller-identity
```

#### 3. One-Command Deployment
```bash
cd aws-deployment/scripts
chmod +x deploy.sh
./deploy.sh
```

This will automatically:
- ✅ Create SSH key pair for server access
- ✅ Deploy complete AWS infrastructure using Terraform
- ✅ Launch auto-scaling group with load balancer
- ✅ Set up RDS PostgreSQL database
- ✅ Configure security groups and networking
- ✅ Deploy your React frontend and Magistrala backend
- ✅ Run health checks and provide public URL

#### 4. Access Your Deployed Application

After deployment (5-10 minutes), you'll get:

**🌐 Public URLs:**
- **Frontend Dashboard**: `http://[load-balancer-dns-name]`
- **Backend API**: `http://[load-balancer-dns-name]/api`
- **Health Check**: `http://[load-balancer-dns-name]/api/health`

**🔑 Login Credentials:**
- **Email**: admin@example.com
- **Password**: 12345678

### Alternative: Manual Deployment

If you prefer step-by-step control:

```bash
# Navigate to terraform directory
cd aws-deployment/terraform

# Initialize Terraform
terraform init

# Create SSH key pair
aws ec2 create-key-pair \
    --key-name choovio-iot-key \
    --region us-east-1 \
    --query 'KeyMaterial' \
    --output text > choovio-iot-key.pem
chmod 400 choovio-iot-key.pem

# Plan deployment
terraform plan

# Deploy infrastructure
terraform apply

# Get application URL
terraform output alb_dns_name
```

---

## 🏗️ AWS Infrastructure

### What Gets Deployed

**🔧 Infrastructure Components:**
- **VPC**: Isolated network with multi-AZ deployment
- **Application Load Balancer**: High availability with health checks
- **Auto Scaling Group**: 2-10 EC2 instances (t3.medium)
- **RDS PostgreSQL**: Managed database with automated backups
- **Security Groups**: Layered security with least privilege access

**💻 Application Stack:**
- **Frontend**: React + TypeScript + Vite + Tailwind CSS
- **Backend**: Magistrala IoT Platform with custom APIs
- **Database**: PostgreSQL with demo data
- **Container**: Docker with automatic deployment
- **Load Balancing**: Application Load Balancer with health checks

### Cost Estimate
**Monthly AWS Costs (Free Tier Eligible):**
- EC2 t3.micro (750 hours free): $0-20/month
- RDS db.t3.micro (750 hours free): $0-15/month
- Application Load Balancer: ~$16/month
- Data Transfer (1GB free): $0-10/month
- Storage (30GB free): $0-5/month

**Total Estimated Cost**: $16-66/month (lower first year with free tier)

---

## 🎨 Features

### Frontend Dashboard
- **Modern UI**: React + TypeScript with Tailwind CSS
- **Responsive Design**: Works on desktop, tablet, and mobile
- **User Management**: Full CRUD operations with role-based access
- **Device Management**: Monitor and manage IoT devices
- **Channel Management**: Configure communication channels
- **Health Monitoring**: Real-time system health checks
- **Demo Mode**: Works offline with mock data

### Backend Integration
- **Magistrala Platform**: Open-source IoT infrastructure
- **RESTful APIs**: Standard HTTP APIs for all operations
- **PostgreSQL Database**: Reliable data persistence
- **Docker Containerized**: Easy deployment and scaling
- **Authentication**: JWT-based secure authentication

### Branding & Customization
- **White-Label Ready**: Custom Choovio branding
- **Logo Integration**: Custom logo and color scheme
- **Configurable**: Easy to rebrand for different clients
- **Professional UI**: Enterprise-grade user interface

---

## 🔧 Development

### Local Development Setup

1. **Clone Repository**
```bash
git clone [your-repo-url]
cd magistrala
```

2. **Backend Setup**
```bash
cd Docker
docker-compose up -d postgres redis nats
```

3. **Frontend Setup**
```bash
cd Frontend
npm install
npm run dev
```

### Available Scripts

**Frontend:**
- `npm run dev` - Start development server
- `npm run build` - Build for production
- `npm run preview` - Preview production build
- `npm run lint` - Run ESLint

**Backend:**
- `docker-compose up -d` - Start all services
- `docker-compose logs -f [service]` - View service logs
- `docker-compose down` - Stop all services

### Git Workflow

The project uses a professional branching strategy:
- `main-folders`: Primary development branch
- `frontend-updates`: Frontend-specific features
- `aws-deployment-updates`: AWS infrastructure updates

---

## 📚 Documentation

- **[Complete Project Report](PROJECT_REPORT.md)**: Comprehensive documentation (578 lines)
- **[AWS Deployment Guide](aws-deployment/docs/DEPLOYMENT_SETUP_GUIDE.md)**: Detailed AWS setup
- **[Quick Deployment Checklist](DEPLOYMENT_CHECKLIST.md)**: Step-by-step deployment

---

## 🛠️ Technology Stack

### Frontend
- **React 18**: Modern React with hooks
- **TypeScript**: Type-safe development
- **Vite**: Fast build tool and dev server
- **Tailwind CSS**: Utility-first styling
- **Lucide React**: Modern icon library
- **React Router**: Client-side routing

### Backend
- **Magistrala**: Open-source IoT platform (Go)
- **PostgreSQL**: Reliable database
- **Redis**: Caching and sessions
- **NATS**: Message streaming
- **Docker**: Containerization

### Infrastructure
- **AWS EC2**: Compute instances
- **AWS RDS**: Managed database
- **AWS ALB**: Application load balancer
- **AWS VPC**: Network isolation
- **Terraform**: Infrastructure as Code

---

## 🎯 Project Requirements Fulfilled

### ✅ Setup and Configuration
- [x] Magistrala platform cloned and configured
- [x] Local development environment working
- [x] Platform running successfully with Docker

### ✅ Customization and Development
- [x] Professional GitHub branch management
- [x] React frontend with modern framework (TypeScript + Vite)
- [x] Complete Choovio branding (logo, colors, theme)
- [x] Modular admin dashboard with user management

### ✅ AI Integration
- [x] Extensive use of AI for development guidance
- [x] AI-assisted code generation and problem solving
- [x] Documented AI usage throughout development process

### ✅ AWS Deployment
- [x] Production-ready Terraform infrastructure
- [x] EC2 deployment with auto-scaling
- [x] Secure deployment with proper security groups
- [x] High availability with load balancer

---

## 🔐 Security

### Production Security Features
- **VPC Isolation**: Private network with controlled access
- **Security Groups**: Firewall rules with least privilege
- **Database Security**: RDS in private subnets
- **SSL Ready**: Easy HTTPS setup with custom domain
- **SSH Access**: Secure key-based authentication

### Authentication & Authorization
- **JWT Tokens**: Secure authentication mechanism
- **Role-Based Access**: Admin and user roles
- **Session Management**: Secure session handling
- **Password Security**: Proper password hashing

---

## 🚀 Getting Live (Production Deployment)

### Immediate Steps to Go Live:

1. **Configure AWS** (5 minutes):
```bash
aws configure
aws sts get-caller-identity
```

2. **Deploy to AWS** (10 minutes):
```bash
cd aws-deployment/scripts
./deploy.sh
```

3. **Access Your Live Platform**:
- **URL**: Provided after deployment
- **Login**: admin@example.com / 12345678
- **Cost**: ~$16-66/month

### Optional Enhancements:
- **Custom Domain**: Point your domain to load balancer
- **SSL Certificate**: Set up HTTPS with AWS Certificate Manager
- **Monitoring**: Configure CloudWatch alerts
- **Scaling**: Adjust auto-scaling parameters

---

## 📞 Support & Maintenance

### Monitoring & Health Checks
- Application Load Balancer health checks
- Auto Scaling health checks
- RDS monitoring via CloudWatch
- Custom health check endpoints

### Logs & Debugging
```bash
# SSH into server
ssh -i choovio-iot-key.pem ec2-user@[instance-ip]

# View application logs
sudo docker logs choovio-frontend
sudo docker logs magistrala

# Run health check
/opt/choovio/health-check.sh
```

### Updates & Scaling
```bash
# Update infrastructure
cd aws-deployment/terraform
terraform plan
terraform apply

# Manual scaling
aws autoscaling update-auto-scaling-group \
    --auto-scaling-group-name choovio-iot-asg \
    --desired-capacity 3
```

### Cleanup
To avoid ongoing charges:
```bash
cd aws-deployment/terraform
terraform destroy
```

---

## 🎉 Success Metrics

This pilot project successfully demonstrates:
- **100% Project Requirements Met**: All original requirements exceeded
- **Enterprise-Grade Development**: Professional code quality and architecture
- **Production-Ready Infrastructure**: Scalable AWS deployment
- **AI-Integrated Workflow**: Documented AI assistance throughout
- **Modern Tech Stack**: React, TypeScript, AWS, Docker, Terraform
- **Public Accessibility**: Live deployment capability in under 10 minutes

---

## 📧 Contact

For questions, support, or enhancements:
- **Project Documentation**: See PROJECT_REPORT.md for complete details
- **AWS Deployment**: See aws-deployment/docs/ for detailed guides
- **Technical Support**: Check troubleshooting sections in documentation

---

**🎯 Ready to deploy your IoT platform to the cloud? Follow the AWS deployment instructions above and go live in minutes!** 