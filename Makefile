.PHONY: help build run stop clean logs swagger test docker-build docker-up docker-down

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the Go application
	go build -o main .
	go build -o publisher ./cmd/publisher

run: ## Run the application locally
	go run main.go

swagger: ## Generate Swagger documentation
	swag init -g main.go

test: ## Run tests
	go test -v ./...

docker-build: ## Build Docker images
	docker-compose build

docker-up: ## Start all services with Docker Compose
	docker-compose up -d

docker-up-build: ## Build and start all services
	docker-compose up --build -d

docker-down: ## Stop all Docker services
	docker-compose down

docker-down-v: ## Stop all services and remove volumes
	docker-compose down -v

docker-logs: ## Show logs from all services
	docker-compose logs -f

docker-logs-backend: ## Show backend logs
	docker logs -f transjakarta-backend

docker-logs-publisher: ## Show publisher logs
	docker logs -f transjakarta-publisher

docker-restart: ## Restart all services
	docker-compose restart

docker-ps: ## Show running containers
	docker-compose ps

clean: ## Clean build artifacts
	rm -f main publisher
	rm -rf docs/

db-connect: ## Connect to PostgreSQL database
	docker exec -it transjakarta-postgres psql -U postgres -d transjakarta_fleet

rabbitmq-ui: ## Open RabbitMQ Management UI (prints URL)
	@echo "RabbitMQ Management UI: http://localhost:15672"
	@echo "Username: guest"
	@echo "Password: guest"

swagger-ui: ## Open Swagger UI (prints URL)
	@echo "Swagger UI: http://localhost:8080/swagger/index.html"

mqtt-subscribe: ## Subscribe to MQTT messages
	docker exec -it transjakarta-mosquitto mosquitto_sub -t "/fleet/vehicle/+/location" -v

install-deps: ## Install Go dependencies
	go mod download
	go mod tidy

install-tools: ## Install development tools
	go install github.com/swaggo/swag/cmd/swag@latest
