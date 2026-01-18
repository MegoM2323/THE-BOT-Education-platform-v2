.PHONY: test test-unit test-integration test-coverage test-all help clean

# Help target
help:
	@echo "Backend Test Commands:"
	@echo "  make test              - Run unit tests"
	@echo "  make test-unit         - Run unit tests (alias for test)"
	@echo "  make test-integration  - Run integration tests"
	@echo "  make test-coverage     - Run all tests with coverage report"
	@echo "  make test-all          - Run all tests (unit + integration)"
	@echo "  make clean             - Clean test cache and artifacts"

# Run unit tests
# Run tests sequentially (-p 1) to avoid connection pool exhaustion
test:
	@echo "Running unit tests (sequential to avoid connection exhaustion)..."
	@cd .. && go test -p 1 -v ./tests/backend/unit/...

# Alias for test
test-unit: test

# Run integration tests
# Run sequentially to avoid connection pool exhaustion
test-integration:
	@echo "Running integration tests..."
	@echo "Note: Integration tests require a running PostgreSQL database"
	@cd .. && go test -p 1 -v ./tests/backend/integration/...

# Run all tests with coverage
# Run sequentially to avoid connection pool exhaustion
test-coverage:
	@echo "Running all tests with coverage..."
	@cd .. && go test -p 1 -v -coverprofile=coverage.out ./tests/backend/...
	@cd .. && go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run all tests
# Run sequentially to avoid connection pool exhaustion
test-all:
	@echo "Running all backend tests (sequential)..."
	@cd .. && go test -p 1 -v ./tests/backend/...

# Clean test artifacts
clean:
	@echo "Cleaning test artifacts..."
	@rm -f ../coverage.out ../coverage.html
	@go clean -testcache
	@echo "Clean complete"
