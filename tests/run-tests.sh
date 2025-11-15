#!/usr/bin/env bash

# run-tests.sh - Main Test Suite Runner
# Runs all tests using Docker Compose by default
# Usage: ./tests/run-tests.sh [local|docker|all]

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Configuration
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TESTS_DIR="$PROJECT_ROOT/tests"
MODE="${1:-docker}"
HOST_UID=$(id -u)
HOST_GID=$(id -g)

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

# Validate required commands based on mode
validate_requirements() {
    local mode=$1
    local failed=0
    
    case "$mode" in
        docker|all)
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
            ;;
    esac
    
    case "$mode" in
        local|all|api|bench)
            # Check Go
            if ! check_command "go" "Install Go: https://go.dev/doc/install"; then
                failed=1
            else
                log_info "Go version: $(go version)"
            fi
            ;;
    esac
    
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

# Cleanup function
cleanup() {
    if [ "$MODE" = "docker" ] || [ "$MODE" = "all" ]; then
        log_info "Cleaning up Docker resources..."
        cd "$PROJECT_ROOT"
        if [ -n "$DOCKER_COMPOSE_CMD" ]; then
            $DOCKER_COMPOSE_CMD -f tests/docker-compose.test.yml down -v 2>/dev/null || true
        fi
    fi
}

trap cleanup EXIT

# Run tests with Docker Compose (default)
run_docker_tests() {
    log_section "Running Tests with Docker Compose"
    
    cd "$PROJECT_ROOT"
    
    # Export UID and GID for docker-compose
    log_info "Running containers as UID=$HOST_UID, GID=$HOST_GID"
    export HOST_UID=$HOST_UID
    export HOST_GID=$HOST_GID
    log_info "Building and starting test services..."
    if $DOCKER_COMPOSE_CMD -f tests/docker-compose.test.yml up --build --abort-on-container-exit; then
        log_success "Docker tests completed successfully"
        exit_code=0
    else
        log_error "Docker tests failed"
        exit_code=1
    fi
    
    # Collect coverage if available
    log_info "Collecting coverage reports..."
    docker cp subconvergo-test:/app/coverage ./coverage 2>/dev/null || true
    
    # Show logs
    log_info "Saving test logs..."
    $DOCKER_COMPOSE_CMD -f tests/docker-compose.test.yml logs > tests/docker-test-logs.txt 2>&1 || true
    
    return $exit_code
}

# Run tests locally (no Docker)
run_local_tests() {
    log_section "Running Tests Locally"
    
    cd "$PROJECT_ROOT"
    
    # Create coverage directory
    mkdir -p coverage
    
    log_info "Running unit tests..."
    if go test -v -race -coverprofile=coverage/coverage.out ./...; then
        log_success "Unit tests passed"
    else
        log_error "Unit tests failed"
        return 1
    fi
    
    log_info "Generating coverage report..."
    go tool cover -html=coverage/coverage.out -o coverage/coverage.html
    go tool cover -func=coverage/coverage.out | tee coverage/coverage.txt
    
    # Extract coverage percentage
    coverage=$(go tool cover -func=coverage/coverage.out | grep total | awk '{print $3}')
    log_success "Total coverage: $coverage"
    
    return 0
}

# Run API tests
run_api_tests() {
    log_section "Running API Tests"
    
    log_info "Starting application with Docker..."
    cd "$PROJECT_ROOT"
    
    # Start only the main service
    docker-compose -f tests/docker-compose.test.yml up -d subconvergo
    
    # Wait for service
    log_info "Waiting for service to be ready..."
    for i in {1..30}; do
        if curl -s http://localhost:25500/version > /dev/null 2>&1; then
            log_info "Service is ready!"
            break
        fi
        if [ $i -eq 30 ]; then
            log_error "Service did not become ready"
            docker-compose -f tests/docker-compose.test.yml down
            return 1
        fi
        sleep 1
    done
    
    # Run API tests
    log_info "Running API test suite..."
    if bash "$TESTS_DIR/test-api.sh"; then
        log_success "API tests passed"
        result=0
    else
        log_error "API tests failed"
        result=1
    fi
    
    # Cleanup
    docker-compose -f tests/docker-compose.test.yml down
    
    return $result
}

# Run benchmarks
run_benchmarks() {
    log_section "Running Benchmarks"
    
    cd "$PROJECT_ROOT"
    
    mkdir -p coverage
    
    if go test -bench=. -benchmem ./... | tee coverage/benchmark.txt; then
        log_success "Benchmarks completed"
        return 0
    else
        log_error "Benchmarks failed"
        return 1
    fi
}

# Generate comprehensive test report
generate_report() {
    log_section "Test Report"
    
    cd "$PROJECT_ROOT"
    
    # Coverage summary
    if [ -f "coverage/coverage.txt" ]; then
        echo "Coverage Summary:"
        tail -n 5 coverage/coverage.txt
        echo ""
    fi
    
    # Docker info
    if [ "$MODE" = "docker" ] || [ "$MODE" = "all" ]; then
        echo "Docker Images:"
        docker images | grep subconvergo || echo "No subconvergo images found"
        echo ""
    fi
    
    # Test artifacts
    echo "Test Artifacts:"
    [ -f "coverage/coverage.html" ] && echo "  ✓ Coverage Report: coverage/coverage.html"
    [ -f "coverage/coverage.txt" ] && echo "  ✓ Coverage Summary: coverage/coverage.txt"
    [ -f "coverage/benchmark.txt" ] && echo "  ✓ Benchmarks: coverage/benchmark.txt"
    [ -f "tests/docker-test-logs.txt" ] && echo "  ✓ Docker Logs: tests/docker-test-logs.txt"
    
    echo ""
    echo "Quick Commands:"
    echo "  View coverage:  open coverage/coverage.html"
    echo "  View logs:      cat tests/docker-test-logs.txt"
    echo "  Run specific:   ./tests/run-tests.sh local"
}

# Print usage
print_usage() {
    echo "Usage: $0 [docker|local|api|bench|all|help]"
    echo ""
    echo "Modes:"
    echo "  docker (default) - Run full test suite with Docker Compose"
    echo "  local            - Run tests locally without Docker"
    echo "  api              - Run API tests only (requires service)"
    echo "  bench            - Run benchmarks only"
    echo "  all              - Run everything (Docker + local + benchmarks)"
    echo "  help             - Show this help message"
    echo ""
    echo "Examples:"
    echo "  ./tests/run-tests.sh          # Run with Docker (recommended)"
    echo "  ./tests/run-tests.sh local    # Run locally"
    echo "  ./tests/run-tests.sh api      # Test API endpoints"
    echo "  ./tests/run-tests.sh all      # Run all tests"
}

# Main execution
main() {
    local exit_code=0
    
    echo "========================================"
    echo "  Subconvergo Test Suite"
    echo "========================================"
    echo "Mode: $MODE"
    echo "Project: $PROJECT_ROOT"
    echo ""
    
    # Validate requirements before running tests
    validate_requirements "$MODE"
    
    case "$MODE" in
        docker)
            run_docker_tests || exit_code=1
            ;;
        local)
            run_local_tests || exit_code=1
            ;;
        api)
            run_api_tests || exit_code=1
            ;;
        bench)
            run_benchmarks || exit_code=1
            ;;
        all)
            log_info "Running comprehensive test suite..."
            run_docker_tests || exit_code=1
            run_local_tests || exit_code=1
            run_benchmarks || exit_code=1
            run_api_tests || exit_code=1
            ;;
        help|--help|-h)
            print_usage
            exit 0
            ;;
        *)
            log_error "Unknown mode: $MODE"
            print_usage
            exit 1
            ;;
    esac
    
    generate_report
    
    echo ""
    if [ $exit_code -eq 0 ]; then
        log_success "All tests passed! ✨"
    else
        log_error "Some tests failed ❌"
    fi
    
    exit $exit_code
}

main "$@"
