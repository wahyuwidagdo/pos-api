
.PHONY: build run test docker-build docker-run

build:
	go build -o bin/pos-api ./cmd/api

run:
	go run ./cmd/api

test:
	go test -v ./... -count=1

docker-build:
	docker build -t pos-api .

docker-run:
	docker-compose up
