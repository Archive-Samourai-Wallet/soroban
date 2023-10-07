# build go app
FROM golang:1.19-alpine3.18 as gobuild

RUN apk --no-cache --update add ca-certificates
RUN apk --no-cache --update add alpine-sdk linux-headers

WORKDIR  /src
COPY go.* /src/
RUN go mod download

RUN mkdir -p /stage
COPY . /src
RUN go build -a -tags netgo -o /stage/soroban-server ./cmd/server


# final image. see docker/tor to build tor base image
FROM samourai-tor

COPY --from=gobuild /stage/soroban-server /usr/local/bin

USER root
RUN addgroup -S soroban && adduser -S -G soroban soroban

USER soroban
ENTRYPOINT ["soroban-server"]
