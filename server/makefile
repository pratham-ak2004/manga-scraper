# Default task
all:run

install:
	@echo "Installing dependencies..."
	@go get --tool github.com/a-h/templ/cmd/templ@latest
	@bun install
	@echo "Dependencies installed."

# Run application from main.go
run:
	@echo "Starting application..."
	@bunx --bun @tailwindcss/cli -i ./views/global.css -o ./public/assets/css/index.css --minify
	@templ generate
	go run cmd/main.go

# Build application
build:
	@echo "Building application..."
	@bunx --bun @tailwindcss/cli -i ./views/global.css -o ./public/assets/css/index.css --minify
	@go tool templ generate
	go build -o bin/app cmd/main.go

# Build and start application
start:
	@echo "Clean Starting application..."
	@bunx --bun @tailwindcss/cli -i ./views/global.css -o ./public/assets/css/index.css --minify
	@go tool templ generate
	go build -o bin/app cmd/main.go
	./bin/app

generate:
	@echo "Generating styles, templates and SQLC..."
	@bunx --bun @tailwindcss/cli -i ./views/global.css -o ./public/assets/css/index.css --minify
	@echo "Style generation complete."
	@go tool templ generate
	@echo "Template generation complete"
	@if ! ls ./bin/sqlc*.wasm 1> /dev/null 2>&1; then \
		echo "No 'sqlc*.wasm' found in ./bin/ folder."; \
		echo "Downloading sqlc..."; \
		if [ ! -d "./bin" ]; then \
			mkdir -p ./bin; \
		fi; \
		curl -sSfL https://github.com/sqlc-dev/sqlc-gen-go/releases/download/v1.4.0/sqlc-gen-go_1.4.0.wasm -o ./bin/sqlc-1.4.0.wasm; \
	fi
	@if ! [ -x "$$(command -v sqlc)" ]; then \
		echo "SQLC is not installed. Please install SQLC and try again."; \
		exit 1; \
	fi
	sqlc generate
	@echo "SQLC Generation complete."

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

start-db:
	@CONTAINER_NAME="manga-scraper-postgres"; \
	if ! [ -x "$$(command -v docker)" ]; then \
		echo -e "Docker is not installed. Please install docker and try again.\nDocker install guide: https://docs.docker.com/engine/install/"; \
		exit 1; \
	fi; \
	if ! docker info > /dev/null 2>&1; then \
		echo "Docker daemon is not running. Please start Docker and try again."; \
		exit 1; \
	fi; \
	if [ "$$(docker ps -q -f name=$$CONTAINER_NAME)" ]; then \
		echo "Database container '$$CONTAINER_NAME' already running"; \
		exit 0; \
	fi; \
	if [ "$$(docker ps -q -a -f name=$$CONTAINER_NAME)" ]; then \
		docker start "$$CONTAINER_NAME"; \
		echo "Existing database container '$$CONTAINER_NAME' started"; \
		exit 0; \
	fi; \
	set -a; \
	source .env; \
	PASSWORD=$$(echo "$$DATABASE_URL" | awk -F':' '{print $$3}' | awk -F'@' '{print $$1}'); \
	USERNAME=$$(echo "$$DATABASE_URL" | awk -F'//' '{print $$2}' | awk -F':' '{print $$1}'); \
	PORT=$$(echo "$$DATABASE_URL" | awk -F':' '{print $$4}' | awk -F'/' '{print $$1}'); \
	DB=$$(echo "$$DATABASE_URL" | awk -F'/' '{print $$4}'); \
	\
	if [ "$$PASSWORD" = "password" ]; then \
		echo "You are using the default database password"; \
		read -p "Should we generate a random password for you? [y/N]: " -r REPLY; \
		if ! [[ $$REPLY =~ ^[Yy]$$ ]]; then \
			echo "Please change the default password in the .env file and try again"; \
			exit 1; \
		fi; \
		PASSWORD=$$(openssl rand -base64 12 | tr '+/' '-_'); \
		sed -i -e "s#:password@#:$$PASSWORD@#" .env; \
	fi; \
	\
	docker run -d \
		--name $$CONTAINER_NAME \
		-e POSTGRES_USER="$$USERNAME" \
		-e POSTGRES_PASSWORD="$$PASSWORD" \
		-e POSTGRES_DB="$$DB" \
		-p "$$PORT":5432 \
		docker.io/postgres && echo "Database container '$$CONTAINER_NAME' was successfully created"