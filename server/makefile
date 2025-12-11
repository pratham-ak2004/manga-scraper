# Default task
all:run

install:
	@echo "Installing dependencies..."
	@echo "Dependencies installed."

# Run application from main.go
run:
	@echo "Starting application..."
	go run cmd/main.go

# Build application
build:
	@echo "Building application..."
	go build -o bin/app cmd/main.go

# Build and start application
start:
	@echo "Clean Starting application..."
	# @bunx --bun @tailwindcss/cli -i ./views/global.css -o ./public/assets/css/index.css --minify
	# @go tool templ generate
	go build -o bin/app cmd/main.go
	./bin/app

generate:
	@echo "Generating styles and templates..."
	# @bunx --bun @tailwindcss/cli -i ./views/global.css -o ./public/assets/css/index.css --minify
	# @go tool templ generate
	@echo "Template generation complete"

# Remove build
clean:
	@echo "Cleaning bin directory..."
	rm -rf bin
	@echo "Clean complete"

tidy:
	@echo "Tidying Go modules..."
	go mod tidy
	@echo "Tidy complete"

# Start development server with hot reload using air
watch:
	@if command -v ./bin/air > /dev/null; then \
	    echo "Watching...";\
	    ./bin/air -c .air.toml -- -h; \
	else \
	    read -p "Go's 'air' is not installed on your machine. Do you want to install it? [Y/n] " choice; \
	    if [ "$$choice" != "n" ] && [ "$$choice" != "N" ]; then \
	        curl -sSfL https://raw.githubusercontent.com/air-verse/air/master/install.sh | sh -s -- -b ./bin; \
	        ./bin/air -c .air.toml -- -h; \
	        echo "Watching...";\
	    else \
	        echo "You chose not to install air. Exiting..."; \
	        exit 1; \
	    fi; \
	fi