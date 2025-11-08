.PHONY: deps build run-repo run-controller clean certs

# Install dependencies
deps:
	go mod tidy
	go mod download

# Generate certificates
certs:
	@if [ ! -f certs/server.crt ]; then \
		echo "Generating self-signed certificates..."; \
		mkdir -p certs; \
		openssl genrsa -out certs/server.key 2048; \
		openssl req -new -x509 -key certs/server.key -out certs/server.crt -days 365 -subj '/C=JP/ST=Tokyo/L=Tokyo/O=Magic-Cylinder/OU=IT Department/CN=localhost'; \
		echo "Certificates generated successfully"; \
	else \
		echo "Certificates already exist"; \
	fi

# Build all binaries
build: deps
	go build -o bin/repository ./cmd/repository
	go build -o bin/controller ./cmd/controller

# Run repository server
run-repo: certs build
	@echo "Starting Repository server on port 8443..."
	./bin/repository

# Run controller server
run-controller: certs build
	@echo "Starting Controller server on port 8444..."
	./bin/controller

# Run both servers (repository first, then controller)
run-all: certs build
	@echo "Starting both servers..."
	@echo "Repository server will start first, then Controller server"
	@(./bin/repository &) && sleep 2 && ./bin/controller

# Clean build artifacts
clean:
	rm -rf bin/
	rm -rf certs/

# Run tests
test:
	go test ./...

# Help
help:
	@echo "Available targets:"
	@echo "  deps          - Install Go dependencies"
	@echo "  certs         - Generate self-signed certificates"
	@echo "  build         - Build all binaries"
	@echo "  run-repo      - Run repository server only"
	@echo "  run-controller - Run controller server only"
	@echo "  run-all       - Run both servers"
	@echo "  clean         - Clean build artifacts and certificates"
	@echo "  test          - Run tests"
	@echo "  help          - Show this help message"