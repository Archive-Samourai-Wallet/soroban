all: docker

docker:
	docker build -t samourai-tor .

docker-static:
	docker build -t samourai-tor-static . -f Dockerfile.static

run:
	docker run samourai-tor

run-static:
	docker run samourai-tor-static

.PHONY: docker run
