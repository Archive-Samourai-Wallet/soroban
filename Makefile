all: soroban

SOROBAN_BUILD_IMAGE=golang:1.15.6-alpine3.12

soroban:
	docker run -ti --rm --name soroban-go-builder -v $$(pwd):/src -w /src ${SOROBAN_BUILD_IMAGE} go build -tags netgo -ldflags="-s -w" -trimpath -o bin/soroban cmd/server/main.go
	cd bin && sha256sum soroban | tee soroban.sum && cd ..

docker:
	docker build -t samourai-soroban .

docker-static:
	docker build -t samourai-soroban-static . -f Dockerfile.static

compose-build:
	docker-compose build

up: compose-build
	docker-compose up -d

down:
	docker-compose down

test:
	docker run -p 6379:6379 --name=redis_test -d redis:5-alpine
	go test -v ./... -count=1 -run=Test
	docker stop redis_test && docker rm redis_test

.PHONY: soroban docker docker-static compose-build up down test
