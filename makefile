start:
	python app.py

worker:
	python app.py --worker

listener:
	python app.py --listener

flower:
	python app.py --flower

start-rabbit:
	@CONTAINER_NAME="manga-scraper-rabbitmq"; \
	if ! [ -x "$$(command -v docker)" ]; then \
		echo -e "Docker is not installed. Please install docker and try again.\nDocker install guide: https://docs.docker.com/engine/install/"; \
		exit 1; \
	fi; \
	if ! docker info > /dev/null 2>&1; then \
		echo "Docker daemon is not running. Please start Docker and try again."; \
		exit 1; \
	fi; \
	if [ "$$(docker ps -q -f name=$$CONTAINER_NAME)" ]; then \
		echo "RabbitMQ container '$$CONTAINER_NAME' already running"; \
		exit 0; \
	fi; \
	if [ "$$(docker ps -q -a -f name=$$CONTAINER_NAME)" ]; then \
		docker start "$$CONTAINER_NAME"; \
		echo "Existing RabbitMQ container '$$CONTAINER_NAME' started"; \
		exit 0; \
	fi; \
	set -a; \
	source .env; \
	PORT=$$(echo "$$RABBITMQ_URL" | awk -F"://" '{print $$2}' | awk -F"@" '{print $$2}' | awk -F":" '{print $$2}' | awk -F"/" '{print $$1}'); \
	PASSWORD=$$(echo "$$RABBITMQ_URL" | awk -F':' '{print $$3}' | awk -F'@' '{print $$1}'); \
	USERNAME=$$(echo "$$RABBITMQ_URL" | awk -F'//' '{print $$2}' | awk -F':' '{print $$1}'); \
	if [ "$$PASSWORD" = "password" ]; then \
		echo "You are using the default password. Please change it in the .env file and try again"; \
		exit 1; \
	fi; \
	docker run -d \
		--name $$CONTAINER_NAME \
		-e RABBITMQ_DEFAULT_USER="$$USERNAME" \
		-e RABBITMQ_DEFAULT_PASS="$$PASSWORD" \
		-p "$$PORT":5672 \
		-p 15672:15672 \
		-v manga-scraper-rabbitmq-data:/var/lib/rabbitmq \
		rabbitmq:4-management && echo "RabbitMQ container '$$CONTAINER_NAME' was successfully created"

start-redis:
	@CONTAINER_NAME="manga-scraper-redis"; \
	if ! [ -x "$$(command -v docker)" ]; then \
		echo -e "Docker is not installed. Please install docker and try again.\nDocker install guide: https://docs.docker.com/engine/install/"; \
		exit 1; \
	fi; \
	if ! docker info > /dev/null 2>&1; then \
		echo "Docker daemon is not running. Please start Docker and try again."; \
		exit 1; \
	fi; \
	if [ "$$(docker ps -q -f name=$$CONTAINER_NAME)" ]; then \
		echo "RabbitMQ container '$$CONTAINER_NAME' already running"; \
		exit 0; \
	fi; \
	if [ "$$(docker ps -q -a -f name=$$CONTAINER_NAME)" ]; then \
		docker start "$$CONTAINER_NAME"; \
		echo "Existing RabbitMQ container '$$CONTAINER_NAME' started"; \
		exit 0; \
	fi; \
	set -a; \
	source .env; \
	PORT=$$(echo "$$REDIS_URL" | awk -F"://" '{print $$2}' | awk -F"@" '{print $$2}' | awk -F":" '{print $$2}' | awk -F"/" '{print $$1}'); \
	PASSWORD=$$(echo "$$REDIS_URL" | awk -F':' '{print $$3}' | awk -F'@' '{print $$1}'); \
	USERNAME=$$(echo "$$REDIS_URL" | awk -F'//' '{print $$2}' | awk -F':' '{print $$1}'); \
	DATABASE=$$(echo "$$REDIS_URL" | awk -F'/' '{print $$4}'); \
	if [ "$$PASSWORD" = "password" ]; then \
		echo "You are using the default password. Please change it in the .env file and try again"; \
		exit 1; \
	fi; \
	docker run -d \
		--name $$CONTAINER_NAME \
		-p "$$PORT":6379 \
		-v manga-scraper-redis-data:/data \
		redis:8.4 redis-server \
		--requirepass "$$PASSWORD" \
		--appendonly yes \
		--user "$$USERNAME" on ">$$PASSWORD" ~* +@all \
		--databases "$$DATABASE" \
		&& echo "Redis container '$$CONTAINER_NAME' was successfully created"