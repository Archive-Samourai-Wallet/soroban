# Samourai tor docker image

Build static tor binaries and create minimal docker image.
Image contains updated ssl certifcates.
Run with user tor with default `/home/tor/.torrc` configuration.

## Build docker image

```bash
make docker
```

## Run docker

```bash
make run
```
