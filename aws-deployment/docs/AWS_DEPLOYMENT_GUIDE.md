# 🚀 Choovio IoT Platform - AWS Deployment Guide

## 📋 Overview

This guide provides comprehensive instructions for deploying the Choovio IoT Platform to AWS using Infrastructure as Code (Terraform) with automated deployment scripts.

## 🏗️ Architecture Overview

### Infrastructure Components

```
Internet Gateway
        │
    Load Balancer (ALB)
        │
    ┌─────────────────────────────────┐
    │        Public Subnets           │
    │  ┌─────────────────────────────┐ │
    │  │     Auto Scaling Group      │ │
    │  │  ┌─────────┐  ┌─────────┐   │ │
    │  │  │  EC2-1  │  │  EC2-2  │   │ │
    │  │  │Frontend │  │Frontend │   │ │
    │  │  │Backend  │  │Backend  │   │ │
    │  │  └─────────┘  └─────────┘   │ │
    │  └─────────────────────────────┘ │
    └─────────────────────────────────┘
                    │
    ┌─────────────────────────────────┐
    │        Private Subnets          │
    │     ┌─────────────────────┐     │
    │     │    RDS PostgreSQL   │     │
    │     │   (Multi-AZ Setup)  │     │
    │     └─────────────────────┘     │
    └─────────────────────────────────┘
```

### Service Architecture

- **Frontend**: React TypeScript application served via Nginx
- **Backend**: Go-based API server with custom Magistrala integration
- **Database**: AWS RDS PostgreSQL with Multi-AZ deployment
- **Load Balancer**: Application Load Balancer with health checks
- **Auto Scaling**: EC2 instances in Auto Scaling Group (1-3 instances)
- **Networking**: VPC with public/private subnets across 2 AZs

## 📋 Prerequisites

### Required Tools

1. **AWS CLI v2**
   ```bash
   # Already installed in previous steps
   aws --version
   ```

2. **Terraform** (v1.0+)
   ```bash
   # Install Terraform
   brew install terraform
   terraform --version
   ```

3. **Git** (for repository access)

### AWS Account Setup

1. **AWS Account** with appropriate permissions
2. **IAM User/Role** with the following policies:
   - EC2FullAccess
   - RDSFullAccess
   - VPCFullAccess
   - IAMReadOnlyAccess
   - ElasticLoadBalancingFullAccess
   - AutoScalingFullAccess

### Required Permissions

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "ec2:*",
                "rds:*",
                "elasticloadbalancing:*",
                "autoscaling:*",
                "iam:ListInstanceProfiles",
                "iam:PassRole"
            ],
            "Resource": "*"
        }
    ]
}
```

## 🔧 Configuration

### Environment Variables

Create a `.env` file in the `aws-deployment` directory:

```bash
# AWS Configuration
AWS_REGION=us-east-1
AWS_PROFILE=default

# Project Configuration
PROJECT_NAME=choovio-iot
ENVIRONMENT=production

# Instance Configuration
INSTANCE_TYPE=t3.medium
KEY_PAIR_NAME=choovio-iot-key

# Database Configuration
DB_INSTANCE_CLASS=db.t3.micro
DB_STORAGE=20
```

### Terraform Variables

You can customize the deployment by modifying variables in `terraform/main.tf`:

```hcl
variable "aws_region" {
  default = "us-east-1"  # Change to your preferred region
}

variable "instance_type" {
  default = "t3.medium"  # Adjust based on requirements
}
```

## 🚀 Deployment Process

### Step 1: Configure AWS Credentials

```bash
# Configure AWS CLI
aws configure
# Enter your AWS Access Key ID, Secret Access Key, Region, and Output format
```

### Step 2: Install Terraform

```bash
# Install Terraform using Homebrew
brew install terraform

# Verify installation
terraform --version
```

### Step 3: Deploy Infrastructure

```bash
# Navigate to deployment directory
cd aws-deployment

# Make deployment script executable
chmod +x scripts/deploy.sh

# Run deployment
./scripts/deploy.sh
```

### Step 4: Manual Deployment (Alternative)

If you prefer manual control:

```bash
# Navigate to terraform directory
cd aws-deployment/terraform

# Initialize Terraform
terraform init

# Plan deployment
terraform plan

# Apply infrastructure
terraform apply
```

## 🔍 Monitoring and Verification

### Health Checks

The deployment includes automatic health checks for:

1. **Load Balancer Health**: ALB target group health
2. **Frontend Availability**: HTTP 200 response on port 5173
3. **Backend API**: Health endpoint on `/api/health`
4. **Database Connectivity**: RDS PostgreSQL connection

### Manual Verification

```bash
# Get load balancer DNS
cd aws-deployment/terraform
ALB_DNS=$(terraform output -raw alb_dns_name)

# Test frontend
curl -I http://$ALB_DNS

# Test backend API
curl http://$ALB_DNS/api/health

# Test login functionality
curl -X POST http://$ALB_DNS/api/tokens \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"admin123"}'
```

## 🔧 Post-Deployment Configuration

### SSL Certificate (Optional)

1. **Request Certificate**: Use AWS Certificate Manager
2. **Update Load Balancer**: Add HTTPS listener
3. **Update Security Groups**: Allow port 443

```bash
# Request SSL certificate
aws acm request-certificate \
  --domain-name your-domain.com \
  --validation-method DNS \
  --region us-east-1
```

### Custom Domain (Optional)

1. **Update Route 53**: Create A record pointing to ALB
2. **Update Application**: Configure custom domain in frontend

### Scaling Configuration

Auto Scaling is configured with:
- **Minimum**: 1 instance
- **Maximum**: 3 instances  
- **Desired**: 1 instance
- **Health Check**: ELB health check
- **Scale-out**: CPU > 70% for 2 minutes
- **Scale-in**: CPU < 30% for 5 minutes

## 🛡️ Security Considerations

### Network Security

- **VPC**: Isolated network environment
- **Security Groups**: Restrictive inbound/outbound rules
- **Private Subnets**: Database isolated from internet
- **NAT Gateway**: Secure outbound internet access

### Application Security

- **Environment Variables**: Sensitive data stored securely
- **Database**: Encrypted at rest and in transit
- **Access Control**: IAM roles and policies
- **SSH Keys**: Secure instance access

### Production Hardening

1. **Remove SSH Access**: Disable port 22 in production
2. **Database Credentials**: Use AWS Secrets Manager
3. **SSL/TLS**: Implement HTTPS everywhere
4. **Monitoring**: Enable CloudWatch logging
5. **Backup**: Configure automated RDS backups

## 📊 Cost Estimation

### Monthly AWS Costs (us-east-1)

| Service | Configuration | Monthly Cost |
|---------|---------------|--------------|
| EC2 (t3.medium) | 1 instance | ~$30 |
| RDS (db.t3.micro) | PostgreSQL | ~$20 |
| ALB | Standard ALB | ~$20 |
| Data Transfer | 10GB/month | ~$1 |
| **Total** | | **~$71/month** |

*Costs may vary by region and actual usage*

## 🔧 Troubleshooting

### Common Issues

#### 1. Deployment Fails

```bash
# Check Terraform logs
terraform plan -detailed-exitcode

# Validate configuration
terraform validate

# Check AWS credentials
aws sts get-caller-identity
```

#### 2. Application Not Accessible

```bash
# Check security groups
aws ec2 describe-security-groups --group-names choovio-iot-*

# Check target group health
aws elbv2 describe-target-health --target-group-arn <target-group-arn>

# SSH into instance to check logs
ssh -i choovio-iot-key.pem ec2-user@<instance-ip>
sudo journalctl -u choovio-backend -f
```

#### 3. Database Connection Issues

```bash
# Check RDS status
aws rds describe-db-instances --db-instance-identifier choovio-iot-db

# Test connectivity from EC2
telnet <rds-endpoint> 5432
```

### Log Locations

- **User Data Logs**: `/var/log/user-data.log`
- **Application Logs**: `/var/log/choovio/`
- **Nginx Logs**: `/var/log/nginx/`
- **System Logs**: `journalctl -u choovio-backend`

## 🔄 Maintenance

### Updates and Patches

1. **OS Updates**: Managed by Auto Scaling Group refresh
2. **Application Updates**: Deploy via new AMI or user data script
3. **Database Maintenance**: Managed by AWS RDS maintenance window

### Backup and Recovery

- **Database**: Automated daily backups (7-day retention)
- **Application**: Stored in Git repository
- **Configuration**: Infrastructure as Code with Terraform

### Scaling Recommendations

#### For Higher Load:
- Increase instance type (t3.large, t3.xlarge)
- Increase Auto Scaling Group max size
- Enable RDS Multi-AZ for high availability
- Consider Redis for session management

#### For Cost Optimization:
- Use t3.micro for development environments
- Enable RDS storage autoscaling
- Implement scheduled scaling for predictable loads

## 📞 Support and Monitoring

### AWS CloudWatch Integration

- **Custom Metrics**: Application performance metrics
- **Alarms**: CPU, memory, disk usage alerts
- **Logs**: Centralized application logging
- **Dashboards**: Real-time monitoring dashboard

### Recommended Alerts

1. **High CPU Usage**: > 80% for 5 minutes
2. **Low Disk Space**: < 10% available
3. **Application Errors**: HTTP 5xx responses
4. **Database Connection Failures**: Connection timeouts

---

## 🎯 Next Steps

After successful deployment:

1. ✅ Test all functionality through the web interface
2. ✅ Configure monitoring and alerting
3. ✅ Set up automated backups
4. ✅ Configure SSL certificate for production
5. ✅ Implement CI/CD pipeline for updates
6. ✅ Configure custom domain
7. ✅ Set up log aggregation and analysis

---

**For support or questions, refer to the main project documentation or AWS support resources.**