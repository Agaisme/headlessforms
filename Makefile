# Variables
BINARY_NAME=server
BUILD_DIR=bin

.PHONY: all build clean run dev docker-build

all: build

# Build the application (Frontend + Backend)
build:
	@echo "Building Frontend..."
	cd web && npm run build
	@echo "Building Backend..."
	go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/server

# Clean build artifacts
clean:
	rm -rf $(BUILD_DIR)
	rm -rf web/build

# Run the server (Requires build first)
run:
	./$(BUILD_DIR)/$(BINARY_NAME)

# Dev mode (Backend only, usually you run frontend separately in dev)
dev:
	go run ./cmd/server

# Build Docker image
docker-build:
	docker build -t headless-form:latest .

# Run tests
test:
	go test -v ./...

# Run Docker container
docker-run:
	docker run -p 8080:8080 -v $(PWD)/data:/data headless-form:latest

