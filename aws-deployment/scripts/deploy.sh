#!/bin/bash

# Choovio IoT Platform - AWS Deployment Script
# This script automates the complete AWS deployment process

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Configuration
PROJECT_NAME="choovio-iot"
AWS_REGION="us-east-1"
KEY_PAIR_NAME="choovio-iot-key"

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."
    
    # Check AWS CLI
    if ! command -v aws &> /dev/null; then
        log_error "AWS CLI is not installed. Please install it first."
        exit 1
    fi
    
    # Check Terraform
    if ! command -v terraform &> /dev/null; then
        log_error "Terraform is not installed. Please install it first."
        exit 1
    fi
    
    # Check AWS credentials
    if ! aws sts get-caller-identity &> /dev/null; then
        log_error "AWS credentials not configured. Please run 'aws configure' first."
        exit 1
    fi
    
    log_success "All prerequisites met!"
}

# Create SSH key pair
create_key_pair() {
    log_info "Creating SSH key pair..."
    
    if aws ec2 describe-key-pairs --key-names "$KEY_PAIR_NAME" --region "$AWS_REGION" &> /dev/null; then
        log_warning "Key pair '$KEY_PAIR_NAME' already exists."
    else
        aws ec2 create-key-pair \
            --key-name "$KEY_PAIR_NAME" \
            --region "$AWS_REGION" \
            --query 'KeyMaterial' \
            --output text > "${KEY_PAIR_NAME}.pem"
        
        chmod 400 "${KEY_PAIR_NAME}.pem"
        log_success "Key pair created and saved as ${KEY_PAIR_NAME}.pem"
    fi
}

# Deploy infrastructure
deploy_infrastructure() {
    log_info "Deploying AWS infrastructure with Terraform..."
    
    cd "$(dirname "$0")/../terraform"
    
    # Initialize Terraform
    terraform init
    
    # Plan deployment
    terraform plan \
        -var="aws_region=${AWS_REGION}" \
        -var="project_name=${PROJECT_NAME}" \
        -var="key_pair_name=${KEY_PAIR_NAME}" \
        -out=tfplan
    
    # Ask for confirmation
    echo ""
    read -p "Do you want to proceed with the deployment? (y/N): " -n 1 -r
    echo
    
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        # Apply infrastructure
        terraform apply tfplan
        
        # Save outputs
        terraform output -json > ../outputs.json
        
        log_success "Infrastructure deployed successfully!"
        
        # Display important information
        ALB_DNS=$(terraform output -raw alb_dns_name)
        log_info "Load Balancer DNS: $ALB_DNS"
        log_info "Application will be available at: http://$ALB_DNS"
        
    else
        log_warning "Deployment cancelled."
        exit 0
    fi
}

# Wait for deployment
wait_for_deployment() {
    log_info "Waiting for application to be ready..."
    
    cd "$(dirname "$0")/../terraform"
    ALB_DNS=$(terraform output -raw alb_dns_name)
    
    echo "Checking application health..."
    TIMEOUT=600  # 10 minutes
    ELAPSED=0
    
    while [ $ELAPSED -lt $TIMEOUT ]; do
        if curl -s -f "http://$ALB_DNS" > /dev/null; then
            log_success "Application is ready!"
            log_success "Frontend: http://$ALB_DNS"
            log_success "Backend API: http://$ALB_DNS/api/"
            return 0
        fi
        
        echo -n "."
        sleep 30
        ELAPSED=$((ELAPSED + 30))
    done
    
    log_error "Application did not become ready within $TIMEOUT seconds"
    return 1
}

# Run health checks
run_health_checks() {
    log_info "Running health checks..."
    
    cd "$(dirname "$0")/../terraform"
    ALB_DNS=$(terraform output -raw alb_dns_name)
    
    # Test frontend
    if curl -s -f "http://$ALB_DNS" > /dev/null; then
        log_success "✓ Frontend is accessible"
    else
        log_error "✗ Frontend is not accessible"
    fi
    
    # Test backend API
    if curl -s -f "http://$ALB_DNS/api/health" > /dev/null; then
        log_success "✓ Backend API is accessible"
    else
        log_error "✗ Backend API is not accessible"
    fi
}

# Display final information
display_final_info() {
    log_info "Deployment Summary"
    echo "===================="
    
    cd "$(dirname "$0")/../terraform"
    ALB_DNS=$(terraform output -raw alb_dns_name)
    
    echo ""
    echo "🎉 Choovio IoT Platform deployed successfully!"
    echo ""
    echo "📱 Frontend Dashboard: http://$ALB_DNS"
    echo "🔧 Backend API: http://$ALB_DNS/api/"
    echo "📊 Health Check: http://$ALB_DNS/api/health"
    echo ""
    echo "🔑 Default Login Credentials:"
    echo "   Email: admin@example.com"
    echo "   Password: admin123"
    echo ""
    echo "📝 SSH Access:"
    echo "   Key file: ${KEY_PAIR_NAME}.pem"
    echo "   Command: ssh -i ${KEY_PAIR_NAME}.pem ec2-user@<instance-ip>"
    echo ""
    echo "💡 Next Steps:"
    echo "   1. Test the application functionality"
    echo "   2. Configure a custom domain (optional)"
    echo "   3. Set up SSL certificate (optional)"
    echo "   4. Configure monitoring and alerts"
    echo ""
}

# Main deployment function
main() {
    echo "🚀 Choovio IoT Platform - AWS Deployment"
    echo "========================================"
    echo ""
    
    check_prerequisites
    create_key_pair
    deploy_infrastructure
    wait_for_deployment
    run_health_checks
    display_final_info
    
    log_success "Deployment completed successfully! 🎉"
}

# Cleanup function
cleanup() {
    log_info "Cleaning up deployment..."
    
    cd "$(dirname "$0")/../terraform"
    
    echo ""
    read -p "Are you sure you want to destroy all AWS resources? (y/N): " -n 1 -r
    echo
    
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        terraform destroy \
            -var="aws_region=${AWS_REGION}" \
            -var="project_name=${PROJECT_NAME}" \
            -var="key_pair_name=${KEY_PAIR_NAME}" \
            -auto-approve
        
        # Remove key pair
        aws ec2 delete-key-pair --key-name "$KEY_PAIR_NAME" --region "$AWS_REGION" || true
        rm -f "${KEY_PAIR_NAME}.pem"
        
        log_success "All resources cleaned up!"
    else
        log_warning "Cleanup cancelled."
    fi
}

# Parse command line arguments
case "${1:-deploy}" in
    deploy)
        main
        ;;
    cleanup|destroy)
        cleanup
        ;;
    health)
        run_health_checks
        ;;
    *)
        echo "Usage: $0 [deploy|cleanup|health]"
        echo "  deploy  - Deploy the infrastructure (default)"
        echo "  cleanup - Destroy all AWS resources"
        echo "  health  - Run health checks"
        exit 1
        ;;
esac 