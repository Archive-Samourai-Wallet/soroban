# Build static tor binaries

# build torproject from source
FROM alpine:3.10 as torproject

ARG TOR_GIT_URL=https://git.torproject.org/tor.git
ARG TOR_VERSION=tor-0.4.2.5

RUN apk --update --no-cache add ca-certificates
RUN apk --no-cache add alpine-sdk automake autoconf
RUN apk --no-cache add openssl-dev libevent-dev zlib-dev

RUN git clone ${TOR_GIT_URL} /tor -b ${TOR_VERSION} --depth 1
    
WORKDIR /tor
RUN ./autogen.sh
RUN ./configure                                           \
    --disable-system-torrc                                \
    --disable-asciidoc                                    \
    --disable-unittests                                   \
    --prefix=/stage
RUN make -j 4 && make install

RUN cp /stage/etc/tor/torrc.sample /stage/.torrc


# final image
FROM alpine:3.10
RUN apk --update --no-cache add ca-certificates
RUN apk --no-cache add openssl libevent zlib

COPY --from=torproject /stage /usr/local
COPY ./proxy_torrc /usr/local/etc/tor/torrc.socks

USER root
RUN addgroup -S tor && adduser -S -G tor tor
RUN mkdir -p /var/lib/tor && chown tor:tor /var/lib/tor

USER tor
RUN cp /usr/local/etc/tor/torrc.sample /home/tor/.torrc

EXPOSE 9150

ENTRYPOINT [ "tor" ]
