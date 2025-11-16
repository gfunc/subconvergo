#!/usr/bin/env bash

# run-tests.sh - Go unit test runner
# Usage: ./tests/run-tests.sh [unit|coverage|bench]

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Configuration
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
MODE="${1:-unit}"

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

# Run unit tests
run_unit_tests() {
    log_section "Running Unit Tests"
    
    cd "$PROJECT_ROOT"
    
    if go test -v -race ./...; then
        log_success "Unit tests passed"
        return 0
    else
        log_error "Unit tests failed"
        return 1
    fi
}

# Run tests with coverage
run_coverage_tests() {
    log_section "Running Tests with Coverage"
    
    cd "$PROJECT_ROOT"
    
    mkdir -p coverage
    
    log_info "Running unit tests with coverage..."
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
    log_info "Coverage report: coverage/coverage.html"
    
    return 0
}

# Run benchmarks
run_benchmarks() {
    log_section "Running Benchmarks"
    
    cd "$PROJECT_ROOT"
    
    mkdir -p coverage
    
    if go test -bench=. -benchmem ./... | tee coverage/benchmark.txt; then
        log_success "Benchmarks completed"
        log_info "Benchmark results: coverage/benchmark.txt"
        return 0
    else
        log_error "Benchmarks failed"
        return 1
    fi
}

# Print usage
print_usage() {
    echo "Usage: $0 [unit|coverage|bench|help]"
    echo ""
    echo "Modes:"
    echo "  unit (default)   - Run unit tests"
    echo "  coverage         - Run tests with coverage report"
    echo "  bench            - Run benchmarks"
    echo "  help             - Show this help message"
    echo ""
    echo "Examples:"
    echo "  ./tests/run-tests.sh          # Run unit tests"
    echo "  ./tests/run-tests.sh coverage # Generate coverage report"
    echo "  ./tests/run-tests.sh bench    # Run benchmarks"
    echo ""
    echo "For smoke tests (integration/API tests):"
    echo "  python -m tests.smoke"
}

# Main execution
main() {
    local exit_code=0
    
    echo "========================================"
    echo "  Subconvergo Go Test Suite"
    echo "========================================"
    echo "Mode: $MODE"
    echo "Project: $PROJECT_ROOT"
    echo ""
    
    # Check Go is available
    if ! command -v go &> /dev/null; then
        log_error "Go not found. Please install Go: https://go.dev/doc/install"
        exit 1
    fi
    
    log_info "Go version: $(go version)"
    echo ""
    
    case "$MODE" in
        unit)
            run_unit_tests || exit_code=1
            ;;
        coverage)
            run_coverage_tests || exit_code=1
            ;;
        bench)
            run_benchmarks || exit_code=1
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
    
    echo ""
    if [ $exit_code -eq 0 ]; then
        log_success "Tests completed successfully! ✨"
    else
        log_error "Tests failed ❌"
    fi
    
    exit $exit_code
}

main "$@"
