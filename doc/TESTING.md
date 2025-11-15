# Testing Guide - Subconvergo

Complete testing documentation for the Subconvergo project.

## Table of Contents

- [Quick Start](#quick-start)
- [Test Infrastructure](#test-infrastructure)
- [Running Tests](#running-tests)
- [Test Coverage](#test-coverage)
- [Command Requirements](#command-requirements)
- [Docker Testing](#docker-testing)
- [API Testing](#api-testing)
- [Troubleshooting](#troubleshooting)

## Quick Start

### Run All Tests (Recommended)

```bash
# Run complete test suite with Docker
./tests/run-tests.sh

# Or using Makefile
make test
```

This single command:
- ✅ Builds and starts Docker containers
- ✅ Runs all unit tests
- ✅ Runs integration tests
- ✅ Generates coverage reports
- ✅ Collects test logs
- ✅ Cleans up automatically

### Quick Status Check

```bash
# Current test status
make test-local

# Expected output:
# ok    github.com/gfunc/subconvergo/parser     coverage: 81.8%
# ok    github.com/gfunc/subconvergo/generator  coverage: 72.0%
# ok    github.com/gfunc/subconvergo/handler    coverage: 30.4%
# ok    github.com/gfunc/subconvergo/tests      [integration]
```

## Test Infrastructure

### Directory Structure

```
tests/
├── run-tests.sh              # Main test runner (Docker-first)
├── test-api.sh               # HTTP endpoint testing
├── test-docker.sh            # Docker-specific testing
├── run-all-tests.sh          # Legacy comprehensive runner
├── docker-compose.test.yml   # Docker Compose orchestration
├── Dockerfile.test           # Test container image
├── integration_test.go       # Integration tests
└── mock-data/               # Mock subscription data
    ├── subscription-ss.txt
    └── clash-subscription.yaml
```

### Test Packages

| Package | Tests | Coverage | Description |
|---------|-------|----------|-------------|
| `parser` | 15 | 81.8% | Subscription parsing, all proxy protocols |
| `generator` | 10 | 72.0% | Config generation for all client formats |
| `handler` | 9 | 30.4% | HTTP handlers, endpoints, CORS |
| `tests` | 3 | N/A | Integration tests, end-to-end workflows |

**Total: 37 tests across 4 packages**

## Running Tests

### Test Modes

```bash
# Docker mode (default) - Full test suite in containers
./tests/run-tests.sh docker
make test

# Local mode - Run tests without Docker (faster for development)
./tests/run-tests.sh local  
make test-local

# API tests only
./tests/run-tests.sh api
make test-api

# Benchmarks only
./tests/run-tests.sh bench
make bench

# Everything - Docker + local + API + benchmarks
./tests/run-tests.sh all
make test-all
```

### Using Makefile

```bash
# Quick test commands
make test              # Run tests with Docker
make test-local        # Run tests locally
make test-unit         # Unit tests only
make test-integration  # Integration tests
make test-api          # API tests
make coverage          # Generate coverage
make bench             # Benchmarks

# Docker commands
make docker-compose-test  # Full Docker Compose test
make docker-build-test    # Build test image

# Code quality
make lint              # Run linter
make fmt               # Format code
make vet               # Run go vet
```

### Running Specific Tests

```bash
# Test specific package
go test ./parser -v
go test ./generator -v
go test ./handler -v

# Run specific test function
go test ./parser -run TestParseClashYAML -v
go test ./generator -run TestGenerateClash -v

# Run with race detection
go test -race ./...

# Run with timeout
go test ./... -timeout 5m

# Bypass test cache
go test ./... -count=1
```

## Test Coverage

### Current Coverage

- **parser**: 81.8% ✅
- **generator**: 72.0% ✅
- **handler**: 30.4% ⚠️
- **Overall**: ~62%

### Generate Coverage Reports

```bash
# HTML coverage report
make coverage
open coverage/coverage.html  # macOS
xdg-open coverage/coverage.html  # Linux

# Terminal coverage
go test ./... -cover

# Coverage for specific package
go test ./parser -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html

# Coverage by function
go test ./parser -coverprofile=coverage.out
go tool cover -func=coverage.out
```

### Coverage Output Files

After running tests, coverage reports are saved to:

- `coverage/coverage.html` - Interactive coverage report
- `coverage/coverage.txt` - Coverage summary
- `coverage/coverage.out` - Raw coverage data
- `coverage/benchmark.txt` - Benchmark results

## Command Requirements

### Docker Mode

**Required commands:**
- `docker` - [Install Docker](https://docs.docker.com/get-docker/)
- `docker-compose` OR `docker compose` - [Install Docker Compose](https://docs.docker.com/compose/install/)

The test scripts automatically detect which Docker Compose format is available:
- **Legacy**: `docker-compose` (standalone binary)
- **Modern**: `docker compose` (Docker CLI plugin)

**Example output:**
```
[INFO] Using Docker Compose command: docker compose
```

### Local Mode

**Required commands:**
- `go` (v1.25.3+) - [Install Go](https://go.dev/doc/install)

**Example output:**
```
[INFO] Go version: go version go1.25.4 linux/amd64
```

### API Mode

**Required commands:**
- `go` - For running tests
- `curl` - For HTTP requests

### Command Detection

All test scripts include automatic command detection with helpful error messages:

```bash
# If docker is missing:
[ERROR] Required command 'docker' not found
Hint: Install Docker: https://docs.docker.com/get-docker/

# If Go is missing:
[ERROR] Required command 'go' not found
Hint: Install Go: https://go.dev/doc/install
```

## Docker Testing

### Docker Compose Testing

```bash
# Run complete Docker Compose test suite
./tests/run-tests.sh docker

# Or manually
docker compose -f tests/docker-compose.test.yml up --build --abort-on-container-exit
docker compose -f tests/docker-compose.test.yml down -v
```

### Docker Services

The Docker Compose setup includes:

| Service | Purpose | Port |
|---------|---------|------|
| `subconvergo` | Main application | 25500 |
| `subconvergo-test` | Unit test runner | - |
| `integration-test` | Integration tests | - |
| `mock-subscription` | Mock subscription server | 8081 |
| `nginx-loadbalancer` | Load balancer for testing | 8888 |

### Docker Test Scripts

```bash
# Build Docker images
./tests/test-docker.sh build

# Run unit tests in Docker
./tests/test-docker.sh unit

# Run integration tests
./tests/test-docker.sh integration

# Test Docker image
./tests/test-docker.sh test-image

# Security scan
./tests/test-docker.sh scan

# Run everything
./tests/test-docker.sh all
```

### Viewing Docker Logs

```bash
# View all logs
docker compose -f tests/docker-compose.test.yml logs

# Follow logs in real-time
docker compose -f tests/docker-compose.test.yml logs -f

# View specific service
docker compose -f tests/docker-compose.test.yml logs subconvergo-test

# Test logs are also saved to:
cat tests/docker-test-logs.txt
```

## API Testing

### Running API Tests

```bash
# Using test script (starts service automatically)
./tests/run-tests.sh api

# Or manually
# Terminal 1: Start service
docker compose -f tests/docker-compose.test.yml up subconvergo

# Terminal 2: Run API tests
./tests/test-api.sh
```

### API Test Coverage

The API test suite (`tests/test-api.sh`) covers:

- ✅ `/version` endpoint
- ✅ `/health` endpoint
- ✅ `/readconf` endpoint
- ✅ `/sub` subscription conversion (all formats)
- ✅ `/getruleset` ruleset fetching
- ✅ CORS headers
- ✅ URL encoding
- ✅ Parameter validation
- ✅ Error handling
- ✅ Response times
- ✅ Concurrent requests (100 simultaneous)

### Custom API URL

```bash
# Test against custom URL
SUBCONVERGO_URL=http://localhost:8080 ./tests/test-api.sh
```

## Troubleshooting

### Common Issues

#### 1. Docker Compose Not Found

**Error:**
```
[ERROR] Docker Compose not found (tried 'docker-compose' and 'docker compose')
```

**Solution:**
```bash
# Install Docker Compose plugin
sudo apt-get install docker-compose-plugin

# Or install standalone
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose
```

#### 2. Permission Denied (Docker)

**Error:**
```
permission denied while trying to connect to the Docker daemon socket
```

**Solution:**
```bash
# Add user to docker group
sudo usermod -aG docker $USER
# Logout and login again
```

#### 3. Port Already in Use

**Error:**
```
bind: address already in use
```

**Solution:**
```bash
# Find and kill process using the port
lsof -ti:25500 | xargs kill -9

# Or change port in docker-compose.test.yml
```

#### 4. Tests Fail on Clean Run

**Solution:**
```bash
# Clean and rebuild
make clean
go mod tidy
go mod download
make test
```

#### 5. Import Errors

**Solution:**
```bash
# Update dependencies
go mod download
go mod verify
go mod tidy
```

### Cleanup

```bash
# Clean test artifacts
make clean

# Or manually
rm -f coverage.out coverage.html *.test
rm -rf coverage/
rm -f tests/docker-test-logs.txt

# Clean Docker resources
docker compose -f tests/docker-compose.test.yml down -v
docker system prune -f

# Remove test images
docker rmi subconvergo:test subconvergo:latest
```

## Testing Workflows

### Daily Development

```bash
# 1. Make code changes
vim parser/parser.go

# 2. Quick local test
go test ./parser -v

# 3. Full test before commit
./tests/run-tests.sh

# 4. View coverage
open coverage/coverage.html
```

### Before Committing

```bash
# Run comprehensive test suite
make test-all

# Check coverage
make coverage

# Lint code
make lint
make fmt
```

### CI/CD Pipeline

```bash
# Simulate CI locally
make ci-docker

# Full CI with all checks
make ci-full
```

### Performance Testing

```bash
# Run benchmarks
./tests/run-tests.sh bench

# Or with make
make bench

# Detailed benchmark
go test ./parser -bench=. -benchmem -benchtime=10s
```

## Test Output

### Test Artifacts

After running tests, the following files are generated:

```
coverage/
├── coverage.html       # Interactive coverage viewer
├── coverage.txt        # Coverage summary
├── coverage.out        # Raw coverage data
└── benchmark.txt       # Benchmark results

tests/
└── docker-test-logs.txt  # Complete Docker test logs

test-results.txt        # Full test output (run-all-tests.sh)
```

### Quick Commands

```bash
# View coverage report
open coverage/coverage.html         # macOS
xdg-open coverage/coverage.html     # Linux

# View test logs
cat tests/docker-test-logs.txt

# View coverage summary
cat coverage/coverage.txt

# View benchmarks
cat coverage/benchmark.txt
```

## Tips & Best Practices

1. **Use Docker for Consistency**: `./tests/run-tests.sh` ensures same environment everywhere
2. **Fast Iteration**: Use `go test ./package` for quick local tests during development
3. **Before Commits**: Always run `make test` or `./tests/run-tests.sh`
4. **Coverage Target**: Aim for >80% coverage per package
5. **Benchmarks**: Run benchmarks on the same hardware for fair comparison
6. **Parallel Tests**: Go runs tests in parallel by default, use `-p 1` to disable
7. **Test Cache**: Use `-count=1` to bypass test cache
8. **Docker Cleanup**: Tests auto-cleanup, but run `docker system prune` if needed
9. **Watch Mode**: Use `entr` or `fswatch` for automatic test runs on file changes
10. **Verbose Output**: Use `-v` flag to see detailed test output

## Quick Reference

```bash
# 🚀 Most Common Commands
./tests/run-tests.sh        # Full test suite with Docker
make test                   # Same as above
make test-local             # Local tests (no Docker)
go test ./...               # Direct Go test

# 🔧 Development
go test ./parser -v         # Test specific package
go test ./... -run TestName # Run specific test
make dev                    # Run in dev mode

# 🐳 Docker
make docker-compose-test    # Full Docker Compose test
docker compose -f tests/docker-compose.test.yml up

# 📊 Reports
make coverage               # Generate coverage
make bench                  # Run benchmarks
cat tests/docker-test-logs.txt  # View logs

# 🧹 Cleanup
make clean                  # Clean artifacts
docker compose -f tests/docker-compose.test.yml down -v
```

## Getting Help

```bash
# View test runner help
./tests/run-tests.sh help

# View test options
go help test

# View benchmark options
go help testflag

# View Makefile targets
make help
```

## Related Documentation

- [QUICKSTART.md](QUICKSTART.md) - Quick start guide
- [DEVELOPMENT.md](DEVELOPMENT.md) - Development guide
- [README.md](README.md) - Project overview

---

**Last Updated**: November 13, 2025  
**Test Status**: ✅ All 37 tests passing  
**Coverage**: 62% overall (parser 81.8%, generator 72.0%, handler 30.4%)
