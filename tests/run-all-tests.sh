#!/usr/bin/env bash

# run-all-tests.sh - Master test runner
# Runs all tests: unit, integration, and API tests

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Configuration
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
COVERAGE_DIR="$PROJECT_ROOT/coverage"
TEST_RESULTS="$PROJECT_ROOT/test-results.txt"

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

# Validate required commands
validate_requirements() {
    local failed=0
    
    # Check Go
    if ! check_command "go" "Install Go: https://go.dev/doc/install"; then
        failed=1
    else
        log_info "Go version: $(go version)"
    fi
    
    # Check curl (for API tests)
    if ! check_command "curl" "Install curl: sudo apt-get install curl (Ubuntu/Debian) or brew install curl (macOS)"; then
        log_warn "curl not found - API tests may fail"
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

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_section() {
    echo ""
    echo "========================================"
    echo -e "${YELLOW}$1${NC}"
    echo "========================================"
}

# Clean previous test results
cleanup() {
    log_info "Cleaning up previous test results..."
    rm -rf "$COVERAGE_DIR"
    mkdir -p "$COVERAGE_DIR"
    rm -f "$TEST_RESULTS"
}

# Run unit tests
run_unit_tests() {
    log_section "Running Unit Tests"
    
    cd "$PROJECT_ROOT"
    
    if go test -v -race -coverprofile="$COVERAGE_DIR/coverage.out" ./... 2>&1 | tee -a "$TEST_RESULTS"; then
        log_success "Unit tests passed"
        
        # Generate coverage report
        log_info "Generating coverage report..."
        go tool cover -html="$COVERAGE_DIR/coverage.out" -o "$COVERAGE_DIR/coverage.html"
        go tool cover -func="$COVERAGE_DIR/coverage.out" | tee "$COVERAGE_DIR/coverage.txt"
        
        # Extract coverage percentage
        coverage=$(go tool cover -func="$COVERAGE_DIR/coverage.out" | grep total | awk '{print $3}')
        log_success "Total coverage: $coverage"
        
        return 0
    else
        log_error "Unit tests failed"
        return 1
    fi
}

# Run benchmarks
run_benchmarks() {
    log_section "Running Benchmarks"
    
    cd "$PROJECT_ROOT"
    
    if go test -bench=. -benchmem ./... 2>&1 | tee "$COVERAGE_DIR/benchmark.txt"; then
        log_success "Benchmarks completed"
        return 0
    else
        log_error "Benchmarks failed"
        return 1
    fi
}

# Run integration tests
run_integration_tests() {
    log_section "Running Integration Tests"
    
    cd "$PROJECT_ROOT"
    
    if [ -d "tests" ]; then
        if go test -v ./tests/... 2>&1 | tee -a "$TEST_RESULTS"; then
            log_success "Integration tests passed"
            return 0
        else
            log_error "Integration tests failed"
            return 1
        fi
    else
        log_info "No integration tests directory found"
        return 0
    fi
}

# Build application
build_app() {
    log_section "Building Application"
    
    cd "$PROJECT_ROOT"
    
    if go build -o "$PROJECT_ROOT/subconvergo" .; then
        log_success "Build successful"
        return 0
    else
        log_error "Build failed"
        return 1
    fi
}

# Run API tests
run_api_tests() {
    log_section "Running API Tests"
    
    # Start the application in background
    log_info "Starting application..."
    "$PROJECT_ROOT/subconvergo" &
    APP_PID=$!
    
    # Wait for application to start
    sleep 3
    
    # Run API tests
    if bash "$PROJECT_ROOT/scripts/test-api.sh"; then
        log_success "API tests passed"
        result=0
    else
        log_error "API tests failed"
        result=1
    fi
    
    # Stop the application
    log_info "Stopping application..."
    kill $APP_PID 2>/dev/null || true
    wait $APP_PID 2>/dev/null || true
    
    return $result
}

# Run linter
run_linter() {
    log_section "Running Linter"
    
    cd "$PROJECT_ROOT"
    
    # Check if golangci-lint is installed
    if command -v golangci-lint &> /dev/null; then
        if golangci-lint run ./... 2>&1 | tee "$COVERAGE_DIR/lint.txt"; then
            log_success "Linter passed"
            return 0
        else
            log_error "Linter found issues"
            return 1
        fi
    else
        log_info "golangci-lint not installed, skipping"
        return 0
    fi
}

# Run security scan
run_security_scan() {
    log_section "Running Security Scan"
    
    cd "$PROJECT_ROOT"
    
    # Check if gosec is installed
    if command -v gosec &> /dev/null; then
        if gosec -fmt=text ./... 2>&1 | tee "$COVERAGE_DIR/security.txt"; then
            log_success "Security scan passed"
            return 0
        else
            log_error "Security issues found"
            return 1
        fi
    else
        log_info "gosec not installed, skipping"
        return 0
    fi
}

# Generate test summary
generate_summary() {
    log_section "Test Summary"
    
    total_tests=0
    passed_tests=0
    failed_tests=0
    
    # Count test results
    if [ -f "$TEST_RESULTS" ]; then
        total_tests=$(grep -c "^=== RUN" "$TEST_RESULTS" || echo "0")
        passed_tests=$(grep -c "^--- PASS" "$TEST_RESULTS" || echo "0")
        failed_tests=$(grep -c "^--- FAIL" "$TEST_RESULTS" || echo "0")
    fi
    
    echo "Total Tests: $total_tests"
    echo -e "Passed: ${GREEN}$passed_tests${NC}"
    echo -e "Failed: ${RED}$failed_tests${NC}"
    echo ""
    
    # Coverage summary
    if [ -f "$COVERAGE_DIR/coverage.txt" ]; then
        echo "Coverage Summary:"
        tail -n 1 "$COVERAGE_DIR/coverage.txt"
        echo ""
    fi
    
    # Output locations
    echo "Test Artifacts:"
    echo "  - Coverage Report: $COVERAGE_DIR/coverage.html"
    echo "  - Coverage Data: $COVERAGE_DIR/coverage.txt"
    echo "  - Benchmark Results: $COVERAGE_DIR/benchmark.txt"
    echo "  - Test Results: $TEST_RESULTS"
    
    if [ -f "$COVERAGE_DIR/lint.txt" ]; then
        echo "  - Lint Results: $COVERAGE_DIR/lint.txt"
    fi
    
    if [ -f "$COVERAGE_DIR/security.txt" ]; then
        echo "  - Security Scan: $COVERAGE_DIR/security.txt"
    fi
}

# Main execution
main() {
    local exit_code=0
    
    echo "========================================"
    echo "  Subconvergo Test Suite"
    echo "========================================"
    echo "Project Root: $PROJECT_ROOT"
    echo ""
    
    # Validate requirements before running tests
    validate_requirements
    
    cleanup
    
    # Run tests in sequence
    run_unit_tests || exit_code=1
    run_benchmarks || true  # Don't fail on benchmark issues
    run_integration_tests || exit_code=1
    run_linter || true  # Don't fail on lint issues
    run_security_scan || true  # Don't fail on security scan issues
    build_app || exit_code=1
    
    # Only run API tests if build succeeded
    if [ $exit_code -eq 0 ]; then
        run_api_tests || exit_code=1
    fi
    
    generate_summary
    
    echo ""
    if [ $exit_code -eq 0 ]; then
        log_success "All tests passed!"
    else
        log_error "Some tests failed"
    fi
    
    exit $exit_code
}

main "$@"
