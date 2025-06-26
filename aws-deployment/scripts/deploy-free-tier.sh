#!/bin/bash

# Choovio IoT Platform - FREE TIER AWS Deployment Script
# Optimized for $0 cost within AWS Free Tier limits

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
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

log_free() {
    echo -e "${PURPLE}[FREE TIER]${NC} $1"
}

# Configuration
PROJECT_NAME="choovio-iot-free"
AWS_REGION="us-east-1"
KEY_PAIR_NAME="choovio-iot-free-key"

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites for FREE TIER deployment..."
    
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
    
    # Check if we're in free tier region
    CURRENT_REGION=$(aws configure get region)
    if [ "$CURRENT_REGION" != "us-east-1" ]; then
        log_warning "You're using region '$CURRENT_REGION'. Free tier is best in 'us-east-1'."
        read -p "Continue anyway? (y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            exit 0
        fi
    fi
    
    log_success "All prerequisites met!"
    log_free "Ready for FREE TIER deployment!"
}

# Display cost information
show_cost_info() {
    echo ""
    echo "💰 FREE TIER COST BREAKDOWN"
    echo "============================"
    echo "✅ EC2 t2.micro instance:        FREE (750 hours/month for 12 months)"
    echo "✅ EBS storage (8GB):             FREE (30GB included in free tier)"
    echo "✅ Elastic IP:                    FREE (1 per region)"
    echo "✅ Local PostgreSQL:              FREE (no RDS charges)"
    echo "✅ Data transfer:                 FREE (15GB/month included)"
    echo "✅ Security Groups:               FREE"
    echo "✅ VPC usage:                     FREE"
    echo ""
    echo "🎯 ESTIMATED MONTHLY COST: $0 - $5"
    echo "   (Only if you exceed free tier limits)"
    echo ""
}

# Create SSH key pair
create_key_pair() {
    log_info "Creating SSH key pair for FREE TIER..."
    
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
        log_free "SSH key creation: FREE!"
    fi
}

# Deploy infrastructure
deploy_infrastructure() {
    log_info "Deploying FREE TIER AWS infrastructure..."
    
    cd "$(dirname "$0")/../terraform"
    
    # Initialize Terraform
    terraform init
    
    # Plan deployment with free tier settings
    terraform plan \
        -var="aws_region=${AWS_REGION}" \
        -var="project_name=${PROJECT_NAME}" \
        -var="key_pair_name=${KEY_PAIR_NAME}" \
        -var="instance_type=t2.micro" \
        -out=tfplan
    
    echo ""
    log_free "📋 FREE TIER DEPLOYMENT SUMMARY"
    echo "================================="
    echo "🖥️  Instance: t2.micro (FREE for 750 hours/month)"
    echo "💾 Storage: 8GB EBS (FREE within 30GB limit)"
    echo "🌐 Network: Default VPC (FREE)"
    echo "📡 IP: Elastic IP (FREE - 1 per region)"
    echo "🗃️  Database: Local PostgreSQL (FREE - no RDS)"
    echo ""
    
    # Ask for confirmation
    read -p "🚀 Deploy FREE TIER infrastructure? (y/N): " -n 1 -r
    echo
    
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        # Apply infrastructure
        terraform apply tfplan
        
        # Save outputs
        terraform output -json > ../outputs.json
        
        log_success "FREE TIER infrastructure deployed successfully!"
        
        # Display important information
        PUBLIC_IP=$(terraform output -raw public_ip)
        log_info "🌐 Public IP: $PUBLIC_IP"
        log_info "📱 Application will be available at: http://$PUBLIC_IP"
        log_free "💰 Current cost: $0 (within free tier)"
        
    else
        log_warning "Deployment cancelled."
        exit 0
    fi
}

# Wait for deployment
wait_for_deployment() {
    log_info "Waiting for FREE TIER application to be ready..."
    
    cd "$(dirname "$0")/../terraform"
    PUBLIC_IP=$(terraform output -raw public_ip)
    
    echo "🔄 Checking application health (this may take 5-10 minutes)..."
    TIMEOUT=900  # 15 minutes for free tier (slower instance)
    ELAPSED=0
    
    while [ $ELAPSED -lt $TIMEOUT ]; do
        if curl -s -f "http://$PUBLIC_IP" > /dev/null; then
            log_success "🎉 Application is ready!"
            log_success "📱 Frontend: http://$PUBLIC_IP"
            log_success "🔧 Backend API: http://$PUBLIC_IP:9000"
            log_free "💰 Running cost: $0 (free tier)"
            return 0
        fi
        
        if [ $((ELAPSED % 60)) -eq 0 ]; then
            echo "⏳ Still waiting... (${ELAPSED}s elapsed)"
        else
            echo -n "."
        fi
        sleep 30
        ELAPSED=$((ELAPSED + 30))
    done
    
    log_error "Application did not become ready within $TIMEOUT seconds"
    log_info "💡 Try checking the EC2 instance manually or running health checks"
    return 1
}

# Run health checks
run_health_checks() {
    log_info "Running FREE TIER health checks..."
    
    cd "$(dirname "$0")/../terraform"
    PUBLIC_IP=$(terraform output -raw public_ip)
    
    echo "🔍 Testing services..."
    
    # Test frontend
    if curl -s -f "http://$PUBLIC_IP" > /dev/null; then
        log_success "✅ Frontend is accessible"
    else
        log_error "❌ Frontend is not accessible"
    fi
    
    # Test backend API
    if curl -s -f "http://$PUBLIC_IP:9000/health" > /dev/null; then
        log_success "✅ Backend API is accessible"
    else
        log_warning "⚠️  Backend API is starting (may take a few more minutes)"
    fi
    
    # Test SSH access
    if ssh -i "${KEY_PAIR_NAME}.pem" -o ConnectTimeout=5 -o StrictHostKeyChecking=no ec2-user@$PUBLIC_IP "echo 'SSH connection successful'" &> /dev/null; then
        log_success "✅ SSH access is working"
    else
        log_warning "⚠️  SSH access not ready yet"
    fi
}

# Display final information
display_final_info() {
    log_info "🎯 FREE TIER DEPLOYMENT SUMMARY"
    echo "=================================="
    
    cd "$(dirname "$0")/../terraform"
    PUBLIC_IP=$(terraform output -raw public_ip)
    
    echo ""
    echo "🎉 Choovio IoT Platform deployed successfully on AWS FREE TIER!"
    echo ""
    echo "🌐 ACCESS INFORMATION:"
    echo "   📱 Frontend Dashboard: http://$PUBLIC_IP"
    echo "   🔧 Backend API: http://$PUBLIC_IP:9000"
    echo "   📊 Health Check: http://$PUBLIC_IP:9000/health"
    echo ""
    echo "🔑 LOGIN CREDENTIALS:"
    echo "   📧 Email: admin@example.com"
    echo "   🔐 Password: admin123"
    echo ""
    echo "🖥️  SSH ACCESS:"
    echo "   🔑 Key file: ${KEY_PAIR_NAME}.pem"
    echo "   💻 Command: ssh -i ${KEY_PAIR_NAME}.pem ec2-user@$PUBLIC_IP"
    echo ""
    echo "💰 COST INFORMATION:"
    echo "   💵 Current cost: $0/month (within free tier)"
    echo "   ⏰ Free tier expires: 12 months from AWS account creation"
    echo "   📈 After free tier: ~$8-15/month for t2.micro"
    echo ""
    echo "🔧 MANAGEMENT COMMANDS (via SSH):"
    echo "   sudo /opt/choovio/manage.sh status   - Check system status"
    echo "   sudo /opt/choovio/manage.sh restart  - Restart all services"
    echo "   sudo /opt/choovio/manage.sh logs     - View application logs"
    echo ""
    echo "🚀 NEXT STEPS:"
    echo "   1. ✅ Test the application functionality"
    echo "   2. 📱 Explore the IoT dashboard features"
    echo "   3. 🔧 Try adding devices and channels"
    echo "   4. 🔒 Consider adding SSL certificate (Let's Encrypt - FREE)"
    echo "   5. 📊 Monitor your AWS usage in the billing dashboard"
    echo ""
    echo "🎯 PILOT PROJECT STATUS: COMPLETE!"
    echo ""
}

# Main deployment function
main() {
    echo "🚀 Choovio IoT Platform - AWS FREE TIER Deployment"
    echo "=================================================="
    echo ""
    
    show_cost_info
    check_prerequisites
    create_key_pair
    deploy_infrastructure
    wait_for_deployment
    run_health_checks
    display_final_info
    
    log_success "🎉 FREE TIER deployment completed successfully!"
    log_free "💰 Total cost: $0 (within AWS free tier limits)"
}

# Cleanup function
cleanup() {
    log_info "🧹 Cleaning up FREE TIER deployment..."
    
    cd "$(dirname "$0")/../terraform"
    
    echo ""
    read -p "⚠️  Are you sure you want to destroy all AWS resources? (y/N): " -n 1 -r
    echo
    
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        terraform destroy \
            -var="aws_region=${AWS_REGION}" \
            -var="project_name=${PROJECT_NAME}" \
            -var="key_pair_name=${KEY_PAIR_NAME}" \
            -var="instance_type=t2.micro" \
            -auto-approve
        
        # Remove key pair
        aws ec2 delete-key-pair --key-name "$KEY_PAIR_NAME" --region "$AWS_REGION" 2>/dev/null || true
        rm -f "${KEY_PAIR_NAME}.pem"
        
        log_success "🧹 All resources cleaned up!"
        log_free "💰 You're back to $0 AWS costs!"
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
    info)
        show_cost_info
        display_final_info
        ;;
    *)
        echo "Usage: $0 [deploy|cleanup|health|info]"
        echo "  deploy  - Deploy the FREE TIER infrastructure (default)"
        echo "  cleanup - Destroy all AWS resources"
        echo "  health  - Run health checks"
        echo "  info    - Show deployment information"
        exit 1
        ;;
esac 