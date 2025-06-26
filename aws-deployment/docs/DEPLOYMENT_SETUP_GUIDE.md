# Choovio IoT Platform - AWS Deployment Setup Guide

## 🚀 **Complete AWS Deployment Guide**

This guide will walk you through deploying your Choovio IoT Platform to AWS, making it publicly accessible.

---

## **Prerequisites**

### 1. **AWS Account Setup**
- Create an AWS account at https://aws.amazon.com/
- Sign up for the free tier (eligible for 12 months)

### 2. **Create IAM User for Deployment**
1. Log into AWS Console → IAM → Users → Create User
2. User name: `choovio-deployer`
3. Attach policies directly:
   - `AmazonEC2FullAccess`
   - `AmazonRDSFullAccess`
   - `AmazonVPCFullAccess`
   - `ElasticLoadBalancingFullAccess`
   - `AutoScalingFullAccess`
4. Create user and download CSV with Access Key ID and Secret

### 3. **Configure AWS CLI**
```bash
aws configure
```
Enter when prompted:
- **AWS Access Key ID**: [Your Access Key from CSV]
- **AWS Secret Access Key**: [Your Secret Key from CSV]
- **Default region**: `us-east-1`
- **Default output format**: `json`

### 4. **Verify Configuration**
```bash
aws sts get-caller-identity
```
Should show your account details.

---

## **Deployment Process**

### **Option 1: Automated Deployment (Recommended)**

Run the automated deployment script:

```bash
cd aws-deployment/scripts
chmod +x deploy.sh
./deploy.sh
```

This will:
- ✅ Check all prerequisites
- ✅ Create SSH key pair for server access
- ✅ Deploy complete AWS infrastructure using Terraform
- ✅ Launch auto-scaling group with load balancer
- ✅ Set up RDS PostgreSQL database
- ✅ Configure security groups and networking
- ✅ Deploy your React frontend and Magistrala backend
- ✅ Run health checks

### **Option 2: Manual Step-by-Step Deployment**

If you prefer manual control:

```bash
# 1. Navigate to terraform directory
cd aws-deployment/terraform

# 2. Initialize Terraform
terraform init

# 3. Create SSH key pair
aws ec2 create-key-pair \
    --key-name choovio-iot-key \
    --region us-east-1 \
    --query 'KeyMaterial' \
    --output text > choovio-iot-key.pem
chmod 400 choovio-iot-key.pem

# 4. Plan deployment
terraform plan

# 5. Deploy infrastructure
terraform apply

# 6. Get application URL
terraform output alb_dns_name
```

---

## **What Gets Deployed**

### **🏗️ Infrastructure Components**

1. **Virtual Private Cloud (VPC)**
   - Isolated network environment
   - Public and private subnets across 2 availability zones
   - Internet gateway for public access

2. **Application Load Balancer (ALB)**
   - Distributes traffic across multiple servers
   - High availability and automatic failover
   - Health checks for automatic recovery

3. **Auto Scaling Group**
   - 2-10 EC2 instances (t3.medium)
   - Automatic scaling based on demand
   - Self-healing infrastructure

4. **RDS PostgreSQL Database**
   - Managed database with automated backups
   - Multi-AZ deployment for high availability
   - 20GB-100GB auto-scaling storage

5. **Security Groups**
   - Layered security with least privilege access
   - HTTPS/HTTP access from internet
   - Database access only from application servers

### **💻 Application Stack**

1. **Frontend**: React + TypeScript + Vite + Tailwind CSS
2. **Backend**: Magistrala IoT Platform
3. **Database**: PostgreSQL with demo data
4. **Container**: Docker with automatic deployment
5. **Load Balancing**: Application Load Balancer with health checks

---

## **Access Your Deployed Application**

After deployment completes (5-10 minutes), you'll get:

### **🌐 Public URLs**
- **Frontend Dashboard**: `http://[load-balancer-dns-name]`
- **Backend API**: `http://[load-balancer-dns-name]/api`
- **Health Check**: `http://[load-balancer-dns-name]/api/health`

### **🔑 Login Credentials**
- **Email**: `admin@example.com`
- **Password**: `12345678`
- **Role**: Admin (Full access)

### **🔧 SSH Server Access**
```bash
ssh -i choovio-iot-key.pem ec2-user@[instance-ip]
```

---

## **Cost Estimate**

### **Monthly AWS Costs (Free Tier Eligible)**
- **EC2 t3.micro (750 hours free)**: $0-20/month
- **RDS db.t3.micro (750 hours free)**: $0-15/month
- **Application Load Balancer**: ~$16/month
- **Data Transfer (1GB free)**: $0-10/month
- **Storage (30GB free)**: $0-5/month

**Total Estimated Cost**: $16-66/month (first year lower with free tier)

---

## **Custom Domain Setup (Optional)**

### 1. **Register Domain**
- Use Route 53, GoDaddy, or any domain registrar
- Example: `yourdomain.com`

### 2. **Configure DNS**
- Create CNAME record pointing to ALB DNS name
- Example: `app.yourdomain.com` → `choovio-alb-123456789.us-east-1.elb.amazonaws.com`

### 3. **SSL Certificate (Optional)**
```bash
# Request ACM certificate
aws acm request-certificate \
    --domain-name yourdomain.com \
    --validation-method DNS \
    --region us-east-1
```

---

## **Monitoring and Maintenance**

### **Health Monitoring**
- Application Load Balancer health checks
- Auto Scaling health checks
- RDS monitoring via CloudWatch

### **Logs Access**
```bash
# SSH into server
ssh -i choovio-iot-key.pem ec2-user@[instance-ip]

# View application logs
sudo docker logs choovio-frontend
sudo docker logs magistrala
```

### **Updates and Scaling**
```bash
# Update infrastructure
terraform plan
terraform apply

# Manual scaling
aws autoscaling update-auto-scaling-group \
    --auto-scaling-group-name choovio-iot-asg \
    --desired-capacity 3
```

---

## **Cleanup/Destruction**

To avoid ongoing charges:

```bash
cd aws-deployment/terraform
terraform destroy
```

This will remove all AWS resources and stop billing.

---

## **Troubleshooting**

### **Common Issues**

1. **"Access Denied" Error**
   - Check IAM permissions
   - Verify AWS credentials: `aws sts get-caller-identity`

2. **"Key Pair Already Exists"**
   - Delete existing key: `aws ec2 delete-key-pair --key-name choovio-iot-key`
   - Or use existing key in terraform variables

3. **Application Not Loading**
   - Check security groups allow HTTP traffic
   - Verify EC2 instances are healthy in target groups
   - Check application logs via SSH

4. **Database Connection Issues**
   - Verify RDS security group allows connections from EC2
   - Check database credentials in user data script

### **Support Resources**
- AWS Documentation: https://docs.aws.amazon.com/
- Terraform AWS Provider: https://registry.terraform.io/providers/hashicorp/aws/
- Magistrala Documentation: https://docs.magistrala.abstractmachines.fr/

---

## Next Steps After Deployment

1. ✅ **Test Application**: Verify all features work in production
2. 🔒 **Security Hardening**: Restrict SSH access to your IP only
3. 📊 **Monitoring**: Set up CloudWatch alerts and dashboards
4. 🔄 **CI/CD Pipeline**: Automate deployments with GitHub Actions
5. 📱 **Mobile Testing**: Test responsive design on various devices
6. 🎨 **Custom Branding**: Further customize for your brand
7. 📈 **Analytics**: Add application analytics and monitoring
8. 🔐 **SSL/HTTPS**: Set up secure connections with custom domain

---

**🎉 Congratulations! Your Choovio IoT Platform is now live and publicly accessible on AWS!** 