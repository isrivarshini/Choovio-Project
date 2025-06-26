#!/bin/bash

echo "🔍 Checking Choovio IoT AWS Deployment Status..."
echo "=================================================="

# Instance details from outputs.json
INSTANCE_IP="54.166.137.2"
INSTANCE_ID="i-0ee9062362909161c"

echo "📍 Instance Details:"
echo "   Instance ID: $INSTANCE_ID"
echo "   Public IP: $INSTANCE_IP"
echo ""

echo "🏃 Checking Instance State..."
INSTANCE_STATE=$(aws ec2 describe-instances --instance-ids $INSTANCE_ID --query 'Reservations[0].Instances[0].State.Name' --output text 2>/dev/null)

if [ "$INSTANCE_STATE" = "running" ]; then
    echo "   ✅ Instance is RUNNING"
elif [ "$INSTANCE_STATE" = "stopped" ]; then
    echo "   🛑 Instance is STOPPED"
elif [ "$INSTANCE_STATE" = "terminated" ]; then
    echo "   ❌ Instance is TERMINATED"
else
    echo "   ❓ Instance state: $INSTANCE_STATE"
fi

echo ""
echo "🌐 Testing Network Connectivity..."

# Test SSH
echo "   Testing SSH connection..."
if timeout 5 ssh -i ../choovio-iot-free-key.pem -o ConnectTimeout=5 -o StrictHostKeyChecking=no ec2-user@$INSTANCE_IP "echo 'SSH OK'" 2>/dev/null; then
    echo "   ✅ SSH connection successful"
else
    echo "   ❌ SSH connection failed"
fi

# Test HTTP
echo "   Testing HTTP (port 80)..."
if timeout 5 curl -s http://$INSTANCE_IP >/dev/null 2>&1; then
    echo "   ✅ HTTP connection successful"
else
    echo "   ❌ HTTP connection failed"
fi

# Test Backend API
echo "   Testing Backend API (port 9000)..."
if timeout 5 curl -s http://$INSTANCE_IP:9000/health >/dev/null 2>&1; then
    echo "   ✅ Backend API accessible"
else
    echo "   ❌ Backend API not accessible"
fi

echo ""
echo "💰 Cost Information:"
echo "   Estimated cost: ~$0-5/month (within free tier)"
echo "   Free tier: 750 hours/month for 12 months"

echo ""
echo "🔧 Management Commands:"
echo "   Start instance:  aws ec2 start-instances --instance-ids $INSTANCE_ID"
echo "   Stop instance:   aws ec2 stop-instances --instance-ids $INSTANCE_ID"
echo "   SSH command:     ssh -i choovio-iot-free-key.pem ec2-user@$INSTANCE_IP"
echo "   Frontend URL:    http://$INSTANCE_IP"
echo "   Backend URL:     http://$INSTANCE_IP:9000"

echo ""
echo "📊 Current Status Summary:"
if [ "$INSTANCE_STATE" = "running" ]; then
    echo "   🟢 Your Choovio IoT Platform is DEPLOYED and should be accessible"
elif [ "$INSTANCE_STATE" = "stopped" ]; then
    echo "   🟡 Your infrastructure is deployed but STOPPED (to save costs)"
    echo "   💡 You can start it anytime with: aws ec2 start-instances --instance-ids $INSTANCE_ID"
else
    echo "   🔴 Your deployment may have issues or been terminated"
fi 