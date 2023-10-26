all: soroban

LC_ALL=C
COMMIT=$(shell git rev-parse HEAD)
VERSION=$(shell git describe --abbrev=6)
DATE=$(shell gdate --utc -d 'today 00:00:00' +'%FT%TZ')

soroban:
	mkdir -p ./bin
	go build -tags netgo -ldflags="-s -w -X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.Date=$(DATE)" -trimpath -o ./bin/soroban-server ./cmd/server
	cd bin && sha256sum soroban-server | tee soroban-server.sum && cd ..

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
