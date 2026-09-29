.PHONY: build test run migrate lint clean docker-build docker-up docker-down

BINARY_NAME=orders
DOCKER_IMAGE=orders-api

build:
	go build -o $(BINARY_NAME) ./cmd/api

test:
	go test ./... -race -cover

run: build
	./$(BINARY_NAME)

migrate: build
	./$(BINARY_NAME) migrate

lint:
	golangci-lint run --timeout=5m

clean:
	rm -f $(BINARY_NAME) coverage.out

docker-build:
	docker build -t $(DOCKER_IMAGE) .

docker-up:
	docker compose up -d

docker-down:
	docker compose down
