#!/usr/bin/env bash

# test-docker.sh - Docker-based testing script
# Builds and tests the application in Docker containers

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

# Command detection
check_command() {
    local cmd=$1
    local install_hint=$2
    
    if ! command -v "$cmd" &> /dev/null; then
        log_error "Required command '$cmd' not found"
        if [ -n "$install_hint" ]; then
            echo -e "${YELLOW}Hint:${NC} $install_hint"
        fi
        return 1
    fi
    return 0
}

# Docker Compose command detection (try both formats)
get_docker_compose_cmd() {
    if command -v docker-compose &> /dev/null; then
        echo "docker-compose"
    elif docker compose version &> /dev/null 2>&1; then
        echo "docker compose"
    else
        return 1
    fi
}

# Validate required commands
validate_requirements() {
    local failed=0
    
    # Check Docker
    if ! check_command "docker" "Install Docker: https://docs.docker.com/get-docker/"; then
        failed=1
    fi
    
    # Check Docker Compose (both formats)
    DOCKER_COMPOSE_CMD=$(get_docker_compose_cmd)
    if [ -z "$DOCKER_COMPOSE_CMD" ]; then
        log_error "Docker Compose not found (tried 'docker-compose' and 'docker compose')"
        echo -e "${YELLOW}Hint:${NC} Install Docker Compose: https://docs.docker.com/compose/install/"
        failed=1
    else
        log_info "Using Docker Compose command: $DOCKER_COMPOSE_CMD"
    fi
    
    if [ $failed -eq 1 ]; then
        log_error "Missing required commands. Please install them and try again."
        exit 1
    fi
}

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_section() {
    echo ""
    echo "========================================"
    echo -e "${YELLOW}$1${NC}"
    echo "========================================"
}

# Clean up function
cleanup() {
    log_info "Cleaning up Docker resources..."
    cd "$PROJECT_ROOT"
    if [ -n "$DOCKER_COMPOSE_CMD" ]; then
        $DOCKER_COMPOSE_CMD -f tests/docker-compose.test.yml down -v 2>/dev/null || true
    fi
}

# Trap cleanup on exit
trap cleanup EXIT

# Build Docker images
build_images() {
    log_section "Building Docker Images"
    
    cd "$PROJECT_ROOT"
    
    log_info "Building test image..."
    if docker build -f tests/Dockerfile.test -t subconvergo:test .; then
        log_success "Test image built successfully"
    else
        log_error "Failed to build test image"
        return 1
    fi
    
    log_info "Building main image..."
    if docker build -t subconvergo:latest .; then
        log_success "Main image built successfully"
    else
        log_error "Failed to build main image"
        return 1
    fi
    
    return 0
}

# Run unit tests in Docker
run_docker_unit_tests() {
    log_section "Running Unit Tests in Docker"
    
    cd "$PROJECT_ROOT"
    
    if docker run --rm \
        -v "$PROJECT_ROOT:/app" \
        -v "$PROJECT_ROOT/coverage:/app/coverage" \
        subconvergo:test \
        sh -c "go test -v -coverprofile=coverage/coverage.out ./... && \
               go tool cover -func=coverage/coverage.out"; then
        log_success "Docker unit tests passed"
        return 0
    else
        log_error "Docker unit tests failed"
        return 1
    fi
}

# Run integration tests with docker-compose
run_docker_integration_tests() {
    log_section "Running Integration Tests with Docker Compose"
    
    cd "$PROJECT_ROOT"
    
    log_info "Starting services..."
    if docker-compose -f docker-compose.test.yml up --build --abort-on-container-exit --exit-code-from integration-test; then
        log_success "Docker integration tests passed"
        result=0
    else
        log_error "Docker integration tests failed"
        result=1
    fi
    
    log_info "Collecting logs..."
    docker-compose -f docker-compose.test.yml logs > docker-test-logs.txt
    
    log_info "Stopping services..."
    docker-compose -f docker-compose.test.yml down -v
    
    return $result
}

# Test Docker image
test_docker_image() {
    log_section "Testing Docker Image"
    
    log_info "Starting container..."
    container_id=$(docker run -d -p 25500:25500 subconvergo:latest)
    
    # Wait for container to start
    log_info "Waiting for container to be ready..."
    sleep 5
    
    # Test health check
    log_info "Testing health check..."
    if curl -s http://localhost:25500/version > /dev/null; then
        log_success "Container is healthy"
        result=0
    else
        log_error "Container health check failed"
        result=1
    fi
    
    # Show logs
    log_info "Container logs:"
    docker logs "$container_id"
    
    # Cleanup
    log_info "Stopping container..."
    docker stop "$container_id" > /dev/null
    docker rm "$container_id" > /dev/null
    
    return $result
}

# Run security scan on Docker image
scan_docker_image() {
    log_section "Scanning Docker Image for Vulnerabilities"
    
    # Check if trivy is installed
    if command -v trivy &> /dev/null; then
        log_info "Running Trivy scan..."
        if trivy image --severity HIGH,CRITICAL subconvergo:latest; then
            log_success "No high/critical vulnerabilities found"
            return 0
        else
            log_error "Vulnerabilities found"
            return 1
        fi
    else
        log_info "Trivy not installed, skipping security scan"
        log_info "Install with: brew install trivy (macOS) or snap install trivy (Linux)"
        return 0
    fi
}

# Test multi-architecture build
test_multi_arch() {
    log_section "Testing Multi-Architecture Build"
    
    cd "$PROJECT_ROOT"
    
    log_info "Building for linux/amd64..."
    if docker buildx build --platform linux/amd64 -t subconvergo:amd64 . --load; then
        log_success "AMD64 build successful"
    else
        log_error "AMD64 build failed"
        return 1
    fi
    
    log_info "Building for linux/arm64..."
    if docker buildx build --platform linux/arm64 -t subconvergo:arm64 . --load 2>/dev/null; then
        log_success "ARM64 build successful"
    else
        log_info "ARM64 build skipped (may require QEMU)"
    fi
    
    return 0
}

# Generate test report
generate_report() {
    log_section "Test Report"
    
    echo "Docker Test Results:"
    echo ""
    
    if [ -f "$PROJECT_ROOT/coverage/coverage.out" ]; then
        echo "Coverage:"
        go tool cover -func="$PROJECT_ROOT/coverage/coverage.out" | tail -n 1
        echo ""
    fi
    
    echo "Docker Images:"
    docker images | grep subconvergo || true
    echo ""
    
    if [ -f "$PROJECT_ROOT/docker-test-logs.txt" ]; then
        echo "Logs saved to: docker-test-logs.txt"
    fi
}

# Main execution
main() {
    local exit_code=0
    
    echo "========================================"
    echo "  Docker Testing Suite"
    echo "========================================"
    echo "Project: $PROJECT_ROOT"
    echo ""
    
    # Validate requirements before running tests
    validate_requirements
    
    # Parse arguments
    case "${1:-all}" in
        build)
            build_images || exit_code=1
            ;;
        unit)
            build_images || exit_code=1
            run_docker_unit_tests || exit_code=1
            ;;
        integration)
            build_images || exit_code=1
            run_docker_integration_tests || exit_code=1
            ;;
        test-image)
            build_images || exit_code=1
            test_docker_image || exit_code=1
            ;;
        scan)
            build_images || exit_code=1
            scan_docker_image || exit_code=1
            ;;
        multi-arch)
            test_multi_arch || exit_code=1
            ;;
        all)
            build_images || exit_code=1
            run_docker_unit_tests || exit_code=1
            test_docker_image || exit_code=1
            scan_docker_image || true  # Don't fail on scan issues
            ;;
        *)
            echo "Usage: $0 {build|unit|integration|test-image|scan|multi-arch|all}"
            exit 1
            ;;
    esac
    
    generate_report
    
    echo ""
    if [ $exit_code -eq 0 ]; then
        log_success "Docker tests completed successfully!"
    else
        log_error "Docker tests failed"
    fi
    
    exit $exit_code
}

main "$@"
