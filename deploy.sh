#!/bin/bash

# Plivo PubSub Gateway Deployment Script
# This script helps deploy the PubSub Gateway service

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🚀 Plivo PubSub Gateway Deployment Script${NC}"
echo "=============================================="

# Check if Docker is available
if command -v docker &> /dev/null; then
    echo -e "${GREEN}✅ Docker is available${NC}"
    DOCKER_AVAILABLE=true
else
    echo -e "${YELLOW}⚠️  Docker not available, will use local Go build${NC}"
    DOCKER_AVAILABLE=false
fi

# Check if Go is available
if command -v go &> /dev/null; then
    echo -e "${GREEN}✅ Go is available${NC}"
    GO_AVAILABLE=true
else
    echo -e "${RED}❌ Go is not available. Please install Go to run locally.${NC}"
    GO_AVAILABLE=false
fi

# Set default values
JWT_SECRET_KEY=${JWT_SECRET_KEY:-"default-secret-key-for-development"}
PORT=${PORT:-"8000"}
DEPLOYMENT_MODE=${1:-"local"}

echo -e "\n${BLUE}📋 Deployment Configuration:${NC}"
echo "JWT_SECRET_KEY: ${JWT_SECRET_KEY:0:20}..."
echo "PORT: $PORT"
echo "Mode: $DEPLOYMENT_MODE"

case $DEPLOYMENT_MODE in
    "docker")
        if [ "$DOCKER_AVAILABLE" = true ]; then
            echo -e "\n${BLUE}🐳 Deploying with Docker...${NC}"
            
            # Build the Docker image
            echo "Building Docker image..."
            docker build -t plivo-pubsub-gateway .
            
            # Stop existing container if running
            docker stop pubsub-gateway 2>/dev/null || true
            docker rm pubsub-gateway 2>/dev/null || true
            
            # Run the container
            echo "Starting container..."
            docker run -d \
                --name pubsub-gateway \
                -p $PORT:8000 \
                -e JWT_SECRET_KEY="$JWT_SECRET_KEY" \
                -e PORT=8000 \
                plivo-pubsub-gateway
            
            echo -e "${GREEN}✅ Container started successfully!${NC}"
            echo "Service URL: http://localhost:$PORT"
            echo "Health Check: http://localhost:$PORT/health"
            
        else
            echo -e "${RED}❌ Docker not available. Please install Docker or use 'local' mode.${NC}"
            exit 1
        fi
        ;;
    
    "docker-compose")
        if [ "$DOCKER_AVAILABLE" = true ]; then
            echo -e "\n${BLUE}🐳 Deploying with Docker Compose...${NC}"
            
            # Set environment variables for docker-compose
            export JWT_SECRET_KEY
            export PORT
            
            # Start services
            docker-compose up -d
            
            echo -e "${GREEN}✅ Services started with Docker Compose!${NC}"
            echo "Service URL: http://localhost:$PORT"
            echo "Health Check: http://localhost:$PORT/health"
            
        else
            echo -e "${RED}❌ Docker not available. Please install Docker or use 'local' mode.${NC}"
            exit 1
        fi
        ;;
    
    "local"|*)
        if [ "$GO_AVAILABLE" = true ]; then
            echo -e "\n${BLUE}🔧 Deploying locally with Go...${NC}"
            
            # Set environment variables
            export JWT_SECRET_KEY
            export PORT
            
            # Change to gateway directory
            cd services/gateway
            
            echo "Starting service..."
            echo "Press Ctrl+C to stop the service"
            
            # Run the service
            go run main.go
            
        else
            echo -e "${RED}❌ Go not available. Please install Go or use Docker mode.${NC}"
            exit 1
        fi
        ;;
esac

echo -e "\n${BLUE}📚 Next Steps:${NC}"
echo "1. Test the service: curl http://localhost:$PORT/health"
echo "2. Register a user: curl -X POST http://localhost:$PORT/users/register -H 'Content-Type: application/json' -d '{\"username\": \"testuser\", \"password\": \"password123\"}'"
echo "3. Test WebSocket: Open websocket-test.html in your browser"
echo "4. Run test suite: ./test-complete-flow.sh"
echo "5. Import Postman collection: Plivo-PubSub-API.postman_collection.json"

echo -e "\n${GREEN}🎉 Deployment completed successfully!${NC}"
