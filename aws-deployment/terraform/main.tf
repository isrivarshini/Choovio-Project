# Choovio IoT Platform - AWS Free Tier Deployment
# Optimized for AWS Free Tier to minimize costs

terraform {
  required_version = ">= 1.0"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

# Configure AWS Provider
provider "aws" {
  region = var.aws_region
}

# Variables - Free Tier Optimized
variable "aws_region" {
  description = "AWS region"
  type        = string
  default     = "us-east-1"
}

variable "project_name" {
  description = "Project name for resource naming"
  type        = string
  default     = "choovio-iot-free"
}

variable "environment" {
  description = "Environment name"
  type        = string
  default     = "development"
}

variable "instance_type" {
  description = "EC2 instance type - FREE TIER"
  type        = string
  default     = "t2.micro"  # Free tier eligible
}

variable "key_pair_name" {
  description = "AWS Key Pair name for EC2 access"
  type        = string
  default     = "choovio-iot-free-key"
}

# Data sources
data "aws_availability_zones" "available" {
  state = "available"
}

# Get default VPC (to save costs)
data "aws_vpc" "default" {
  default = true
}

# Get default subnets (to save costs)
data "aws_subnets" "default" {
  filter {
    name   = "vpc-id"
    values = [data.aws_vpc.default.id]
  }
}

# Security Group for EC2 Instance - Simplified
resource "aws_security_group" "web" {
  name_prefix = "${var.project_name}-web-"
  vpc_id      = data.aws_vpc.default.id

  # SSH access
  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  # HTTP access for frontend
  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  # HTTPS access (future)
  ingress {
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  # Frontend dev port
  ingress {
    from_port   = 5173
    to_port     = 5173
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  # Backend API
  ingress {
    from_port   = 9000
    to_port     = 9000
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  # Magistrala services
  ingress {
    from_port   = 8000
    to_port     = 8999
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  # PostgreSQL for local access
  ingress {
    from_port   = 5432
    to_port     = 5432
    protocol    = "tcp"
    cidr_blocks = ["10.0.0.0/8"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name        = "${var.project_name}-web-sg"
    Environment = var.environment
    Tier        = "Free"
  }
}

# FREE TIER: Single EC2 Instance (instead of ALB + ASG)
resource "aws_instance" "choovio_server" {
  ami           = "ami-0c02fb55956c7d316"  # Amazon Linux 2 AMI (Free tier eligible)
  instance_type = var.instance_type
  key_name      = var.key_pair_name

  vpc_security_group_ids = [aws_security_group.web.id]
  
  # Use first available subnet from default VPC
  subnet_id                   = tolist(data.aws_subnets.default.ids)[0]
  associate_public_ip_address = true

  # FREE TIER: Use local SQLite instead of RDS to save costs
  user_data = base64encode(templatefile("${path.module}/user_data.sh", {
    project_name = var.project_name
    aws_region   = var.aws_region
  }))

  root_block_device {
    volume_type = "gp2"
    volume_size = 8  # Free tier allows up to 30GB, using 8GB to be conservative
    encrypted   = false  # Encryption not included in free tier
  }

  tags = {
    Name        = "${var.project_name}-server"
    Environment = var.environment
    Tier        = "Free"
    Project     = "Choovio IoT Platform"
  }
}

# FREE TIER: Elastic IP (one free static IP per region)
resource "aws_eip" "choovio_ip" {
  instance = aws_instance.choovio_server.id
  domain   = "vpc"

  tags = {
    Name        = "${var.project_name}-eip"
    Environment = var.environment
    Tier        = "Free"
  }
}

# Outputs
output "public_ip" {
  description = "Public IP of the server"
  value       = aws_eip.choovio_ip.public_ip
}

output "public_dns" {
  description = "Public DNS name of the server"
  value       = aws_instance.choovio_server.public_dns
}

output "ssh_command" {
  description = "SSH command to connect to the server"
  value       = "ssh -i ${var.key_pair_name}.pem ec2-user@${aws_eip.choovio_ip.public_ip}"
}

output "frontend_url" {
  description = "Frontend application URL"
  value       = "http://${aws_eip.choovio_ip.public_ip}"
}

output "backend_url" {
  description = "Backend API URL"
  value       = "http://${aws_eip.choovio_ip.public_ip}:9000"
}

output "cost_estimate" {
  description = "Monthly cost estimate"
  value       = "~$0-5/month (within free tier limits for first 12 months)"
}

# Security note output
output "security_note" {
  description = "Security consideration"
  value       = "⚠️  This is a development setup. For production, implement proper security groups, SSL certificates, and database security."
} 