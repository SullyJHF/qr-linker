#!/bin/bash

# QR Linker Production Deployment Script
# This script helps deploy the QR Linker application to production with Traefik

set -e

# Function to use the correct docker compose command
docker_compose() {
    local compose_file=""
    if [[ "${DEPLOY_MODE:-}" == "local" ]]; then
        compose_file="-f docker-compose.local.yml"
    fi
    
    if command -v docker-compose &> /dev/null; then
        docker-compose $compose_file "$@"
    else
        docker compose $compose_file "$@"
    fi
}

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Helper functions
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

# Check if Docker and Docker Compose are installed
check_dependencies() {
    log_info "Checking dependencies..."
    
    if ! command -v docker &> /dev/null; then
        log_error "Docker is not installed. Please install Docker first."
        exit 1
    fi
    
    # Check for both docker-compose and docker compose
    if command -v docker-compose &> /dev/null; then
        log_info "Found docker-compose command"
    elif docker compose version &> /dev/null; then
        log_info "Found docker compose plugin"
    else
        log_error "Docker Compose is not installed. Please install Docker Compose first."
        log_info "Either install docker-compose or ensure docker compose plugin is available"
        exit 1
    fi
    
    log_success "Dependencies check passed"
}

# Check environment configuration
check_environment() {
    log_info "Checking environment configuration..."
    
    if [ ! -f .env ]; then
        log_error ".env file not found. Please create one from .env.example"
        log_info "Run: cp .env.example .env"
        log_info "Then edit .env with your configuration"
        exit 1
    fi
    
    # Source the .env file to check variables
    source .env
    
    # Skip Traefik checks for local deployment
    if [[ "${DEPLOY_MODE:-}" != "local" ]]; then
        if [ -z "$TRAEFIK_DOMAIN" ]; then
            log_error "TRAEFIK_DOMAIN is not set in .env file"
            exit 1
        fi
        
        if [ -z "$TRAEFIK_CERT_RESOLVER" ]; then
            log_error "TRAEFIK_CERT_RESOLVER is not set in .env file"
            exit 1
        fi
        
        log_info "Domain: $TRAEFIK_DOMAIN"
        log_info "Cert Resolver: $TRAEFIK_CERT_RESOLVER"
    fi
    
    if [ -z "$BASE_URL" ]; then
        log_error "BASE_URL is not set in .env file"
        exit 1
    fi
    
    log_success "Environment configuration check passed"
    log_info "Base URL: $BASE_URL"
}

# Check if Traefik network exists (only for production)
check_traefik_network() {
    if [[ "${DEPLOY_MODE:-}" == "local" ]]; then
        log_info "Skipping Traefik network check for local deployment"
        return
    fi
    
    log_info "Checking Traefik network..."
    
    if ! docker network ls | grep -q "traefik"; then
        log_error "Traefik network not found. Please ensure Traefik is running."
        log_info "Create the network with: docker network create traefik"
        exit 1
    fi
    
    log_success "Traefik network found"
}

# Build and deploy the application
deploy() {
    if [[ "${DEPLOY_MODE:-}" == "local" ]]; then
        log_info "Starting QR Linker local deployment..."
    else
        log_info "Starting QR Linker production deployment..."
    fi
    
    # Set build version information
    log_info "Setting build version information..."
    
    # Get version info
    BUILD_HASH="${BUILD_HASH:-$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")}"
    BUILD_BRANCH="${BUILD_BRANCH:-$(git branch --show-current 2>/dev/null || echo "unknown")}"
    BUILD_TIME="${BUILD_TIME:-$(date -u +%Y-%m-%dT%H:%M:%SZ)}"
    
    # Export for docker-compose
    export BUILD_HASH="$BUILD_HASH"
    export BUILD_BRANCH="$BUILD_BRANCH"
    export BUILD_TIME="$BUILD_TIME"
    
    log_info "Building commit: $BUILD_HASH from branch $BUILD_BRANCH at $BUILD_TIME"
    
    # Stop existing containers if running
    log_info "Stopping existing containers..."
    docker_compose down --remove-orphans || true
    
    # Build and start containers
    log_info "Building and starting containers..."
    docker_compose up -d --build
    
    # Wait for containers to be healthy
    log_info "Waiting for containers to be healthy..."
    sleep 15
    
    # Health check
    log_info "Performing health check..."
    source .env
    
    if [[ "${DEPLOY_MODE:-}" == "local" ]]; then
        HEALTH_URL="http://localhost:${PORT:-8080}/login"
    else
        HEALTH_URL="https://${TRAEFIK_DOMAIN}/login"
    fi
    
    for i in {1..30}; do
        if curl -f -k "$HEALTH_URL" > /dev/null 2>&1; then
            log_success "Health check passed"
            break
        fi
        if [ $i -eq 30 ]; then
            log_error "Health check failed after 30 attempts"
            log_error "Tried URL: $HEALTH_URL"
            log_error "Deployment may have issues. Check container logs:"
            docker_compose logs --tail=50
            exit 1
        fi
        log_info "Waiting for application to start... ($i/30)"
        sleep 2
    done
    
    # Check container status
    if docker_compose ps | grep -q "Up"; then
        if [[ "${DEPLOY_MODE:-}" == "local" ]]; then
            log_success "QR Linker local deployment successful!"
            log_info "Application should be available at: http://localhost:${PORT:-8080}"
        else
            log_success "QR Linker production deployment successful!"
            log_info "Application should be available at: https://${TRAEFIK_DOMAIN}"
        fi
        
        # Show running containers
        log_info "Running containers:"
        docker_compose ps
        
        # Show next steps
        log_info "Next steps:"
        log_info "1. Create admin user: ./deploy.sh adduser"
        log_info "2. Manage users: ./deploy.sh manage-users"
        log_info "3. View logs: ./deploy.sh logs"
    else
        log_error "Deployment failed. Check container logs:"
        docker_compose logs
        exit 1
    fi
}

# Add a user
add_user() {
    log_info "Adding a new user..."
    docker_compose exec qr-linker ./adduser
}

# Manage users
manage_users() {
    log_info "Opening user management interface..."
    docker_compose exec qr-linker ./manageusers
}

# Show logs
show_logs() {
    log_info "Showing application logs..."
    docker_compose logs -f
}

# Stop the application
stop() {
    log_info "Stopping QR Linker..."
    docker_compose down
    log_success "QR Linker stopped"
}

# Restart the application
restart() {
    log_info "Restarting QR Linker..."
    docker_compose restart
    log_success "QR Linker restarted"
}

# Health check
health_check() {
    log_info "Performing health check..."
    source .env
    
    if [[ "${DEPLOY_MODE:-}" == "local" ]]; then
        HEALTH_URL="http://localhost:${PORT:-8080}/login"
    else
        HEALTH_URL="https://${TRAEFIK_DOMAIN}/login"
    fi
    
    if curl -f -k "$HEALTH_URL" > /dev/null 2>&1; then
        log_success "QR Linker is healthy"
        log_info "URL: $HEALTH_URL"
        return 0
    else
        log_error "QR Linker health check failed"
        log_info "Tried URL: $HEALTH_URL"
        return 1
    fi
}

# Show application status
status() {
    log_info "QR Linker status:"
    docker_compose ps
}

# Show help
show_help() {
    echo "QR Linker Deployment Script"
    echo
    echo "Usage: $0 [COMMAND]"
    echo
    echo "Commands:"
    echo "  deploy         Build and deploy to production (default)"
    echo "  local          Build and deploy locally (no Traefik)"
    echo "  adduser        Add a new user interactively"
    echo "  manage-users   Open user management interface"
    echo "  logs           Show application logs"
    echo "  stop           Stop application"
    echo "  restart        Restart application"
    echo "  status         Show container status"
    echo "  health         Check application health"
    echo "  help           Show this help message"
    echo
    echo "Examples:"
    echo "  $0 deploy         # Deploy to production"
    echo "  $0 local          # Deploy locally for testing"
    echo "  $0 adduser        # Add a new admin user"
    echo "  $0 logs           # Show logs"
    echo "  $0 health         # Check if application is healthy"
    echo "  $0 stop           # Stop containers"
    echo
    echo "Prerequisites:"
    echo "  - Docker and Docker Compose installed"
    echo "  - Traefik running with 'traefik' network"
    echo "  - .env file configured (copy from .env.example)"
}

# Main script logic
main() {
    case "${1:-deploy}" in
        "deploy")
            check_dependencies
            check_environment
            check_traefik_network
            deploy
            ;;
        "local")
            export DEPLOY_MODE="local"
            check_dependencies
            check_environment
            check_traefik_network
            deploy
            ;;
        "adduser")
            add_user
            ;;
        "manage-users")
            manage_users
            ;;
        "logs")
            show_logs
            ;;
        "stop")
            stop
            ;;
        "restart")
            restart
            ;;
        "status")
            status
            ;;
        "health")
            health_check
            ;;
        "help"|"-h"|"--help")
            show_help
            ;;
        *)
            log_error "Unknown command: $1"
            show_help
            exit 1
            ;;
    esac
}

# Run main function with all arguments
main "$@"