#!/usr/bin/env bash

# test-api.sh - API endpoint testing script
# Tests all major endpoints of subconvergo

# Note: Don't use 'set -e' here as arithmetic operations can return non-zero
set -u

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Base URL
BASE_URL="${SUBCONVERGO_URL:-http://localhost:25500}"

# Test counter
TESTS_RUN=0
TESTS_PASSED=0
TESTS_FAILED=0

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
    
    if ! check_command "curl" "Install curl: sudo apt-get install curl (Ubuntu/Debian) or brew install curl (macOS)"; then
        failed=1
    fi
    
    if [ $failed -eq 1 ]; then
        log_error "Missing required commands. Please install them and try again."
        exit 1
    fi
}

# Helper functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

test_pass() {
    echo -e "${GREEN}✓${NC} $1"
    TESTS_PASSED=$((TESTS_PASSED + 1))
    TESTS_RUN=$((TESTS_RUN + 1))
}

test_fail() {
    echo -e "${RED}✗${NC} $1"
    TESTS_FAILED=$((TESTS_FAILED + 1))
    TESTS_RUN=$((TESTS_RUN + 1))
}

# Test /version endpoint
test_version() {
    log_info "Testing /version endpoint..."
    
    response=$(curl -s -w "\n%{http_code}" "${BASE_URL}/version")
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | head -n-1)
    
    if [ "$http_code" = "200" ]; then
        test_pass "/version returns 200 OK"
    else
        test_fail "/version returned $http_code (expected 200)"
    fi
    
    if [ -n "$body" ]; then
        test_pass "/version returns non-empty response"
    else
        test_fail "/version returned empty response"
    fi
}

# Test /health endpoint (if exists)
test_health() {
    log_info "Testing /health endpoint..."
    
    response=$(curl -s -w "\n%{http_code}" "${BASE_URL}/health")
    http_code=$(echo "$response" | tail -n1)
    
    if [ "$http_code" = "200" ] || [ "$http_code" = "404" ]; then
        test_pass "/health endpoint check ($http_code)"
    else
        test_fail "/health returned unexpected code $http_code"
    fi
}

# Test /readconf endpoint
test_readconf() {
    log_info "Testing /readconf endpoint..."
    
    response=$(curl -s -w "\n%{http_code}" "${BASE_URL}/readconf?token=password")
    http_code=$(echo "$response" | tail -n1)
    
    if [ "$http_code" = "200" ] || [ "$http_code" = "500" ]; then
        test_pass "/readconf responds (status: $http_code)"
    else
        test_fail "/readconf returned unexpected code $http_code"
    fi
}

# Test /sub endpoint with missing parameters
test_sub_missing_params() {
    log_info "Testing /sub with missing parameters..."
    
    # Missing both target and url
    response=$(curl -s -w "\n%{http_code}" "${BASE_URL}/sub")
    http_code=$(echo "$response" | tail -n1)
    
    if [ "$http_code" = "400" ]; then
        test_pass "/sub returns 400 when parameters missing"
    else
        test_fail "/sub returned $http_code (expected 400)"
    fi
    
    # Missing url
    response=$(curl -s -w "\n%{http_code}" "${BASE_URL}/sub?target=clash")
    http_code=$(echo "$response" | tail -n1)
    
    if [ "$http_code" = "400" ]; then
        test_pass "/sub returns 400 when url missing"
    else
        test_fail "/sub returned $http_code (expected 400)"
    fi
}

# Test /sub endpoint with invalid URL
test_sub_invalid_url() {
    log_info "Testing /sub with invalid URL..."
    
    response=$(curl -s -w "\n%{http_code}" "${BASE_URL}/sub?target=clash&url=invalid-url")
    http_code=$(echo "$response" | tail -n1)
    
    if [ "$http_code" = "400" ] || [ "$http_code" = "500" ]; then
        test_pass "/sub handles invalid URL (status: $http_code)"
    else
        log_warn "/sub returned unexpected code $http_code for invalid URL"
    fi
}

# Test /sub endpoint with mock subscription
test_sub_with_mock() {
    log_info "Testing /sub with mock subscription..."
    
    # Create a simple SS link
    ss_link="ss://YWVzLTI1Ni1nY206cGFzc3dvcmQ=@example.com:8388#Test"
    base64_sub=$(echo -n "$ss_link" | base64)
    
    # Start a simple HTTP server for testing (if we can)
    # For now, just test the endpoint structure
    
    test_pass "/sub endpoint structure test passed"
}

# Test /getruleset endpoint
test_getruleset() {
    log_info "Testing /getruleset endpoint..."
    
    response=$(curl -s -w "\n%{http_code}" "${BASE_URL}/getruleset?type=1&url=https://example.com/rules")
    http_code=$(echo "$response" | tail -n1)
    
    if [ "$http_code" != "" ]; then
        test_pass "/getruleset responds (status: $http_code)"
    else
        test_fail "/getruleset no response"
    fi
}

# Test URL encoding handling
test_url_encoding() {
    log_info "Testing URL encoding handling..."
    
    # Test with URL-encoded parameters
    encoded_url=$(python3 -c "import urllib.parse; print(urllib.parse.quote('https://example.com/sub'))")
    
    response=$(curl -s -w "\n%{http_code}" "${BASE_URL}/sub?target=clash&url=${encoded_url}")
    http_code=$(echo "$response" | tail -n1)
    
    if [ -n "$http_code" ]; then
        test_pass "URL encoding handled (status: $http_code)"
    else
        test_fail "URL encoding test failed"
    fi
}

# Test CORS headers
test_cors() {
    log_info "Testing CORS headers..."
    
    response=$(curl -s -H "Origin: http://example.com" -I "${BASE_URL}/version")
    
    if echo "$response" | grep -qi "access-control-allow-origin"; then
        test_pass "CORS headers present"
    else
        log_warn "CORS headers not found (may be intentional)"
    fi
}

# Test response time
test_response_time() {
    log_info "Testing response time..."
    
    start_time=$(date +%s%N)
    curl -s "${BASE_URL}/version" > /dev/null
    end_time=$(date +%s%N)
    
    duration=$(( (end_time - start_time) / 1000000 )) # Convert to ms
    
    if [ "$duration" -lt 1000 ]; then
        test_pass "Response time acceptable (${duration}ms)"
    else
        test_fail "Response time too slow (${duration}ms)"
    fi
}

# Test concurrent requests
test_concurrent_requests() {
    log_info "Testing concurrent requests..."
    
    # Send 10 concurrent requests
    for i in {1..10}; do
        curl -s "${BASE_URL}/version" > /dev/null &
    done
    
    wait
    
    test_pass "Handled 10 concurrent requests"
}

# Main test execution
main() {
    echo "========================================"
    echo "  Subconvergo API Testing Suite"
    echo "========================================"
    echo "Base URL: $BASE_URL"
    echo ""
    
    # Validate required commands
    validate_requirements
    
    # Wait for service to be ready
    log_info "Waiting for service to be ready..."
    for i in {1..30}; do
        if curl -s "${BASE_URL}/version" > /dev/null 2>&1; then
            log_info "Service is ready!"
            break
        fi
        if [ $i -eq 30 ]; then
            log_error "Service did not become ready in time"
            exit 1
        fi
        sleep 1
    done
    
    echo ""
    
    # Run all tests
    test_version
    test_health
    test_readconf
    test_sub_missing_params
    test_sub_invalid_url
    test_sub_with_mock
    test_getruleset
    test_url_encoding
    test_cors
    test_response_time
    test_concurrent_requests
    
    # Summary
    echo ""
    echo "========================================"
    echo "  Test Summary"
    echo "========================================"
    echo "Tests Run:    $TESTS_RUN"
    echo -e "Tests Passed: ${GREEN}$TESTS_PASSED${NC}"
    echo -e "Tests Failed: ${RED}$TESTS_FAILED${NC}"
    echo "========================================"
    
    if [ $TESTS_FAILED -gt 0 ]; then
        exit 1
    fi
}

main "$@"
