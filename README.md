# Soroban server in go

## Reproducible build

binary & sha256sum of soroban will be written in `bin` directory

```bash
make soroban
```

## Usage

```bash
  -directoryHostname string
        Directory host
  -directoryPort int
        Directory host
  -directoryType string
        Directory Type (default, redis, memory)
  -domain string
        Directory Domain
  -export string
        Export hidden service secret key from seed to file
  -hostname string
        server address (default localhost) (default "localhost")
  -log string
        Log level (default info) (default "info")
  -p2pBootstrap string
        P2P bootstrap
  -p2pListenPort int
        P2P Listen Port (default 1042)
  -p2pRoom string
        P2P Room (default "samourai")
  -p2pSeed string
        P2P Onion private key seed (default "auto")
  -port int
        Server port (default 4242) (default 4242)
  -prefix string
        Generate Onion with prefix
  -seed string
        Onion private key seed
  -withTor
        Hidden service enabled (default false)
```

## Confidential keys

Configuration file can be use to list confidential keys.
Key prefix is used to find confidential key.
First matching rule is applied.
see [soroban.yml](soroban.yml)

Confidential keys can be:

- Confidential: Every body can add a key. Read must be signed by private key. 
- Readonly: Add or delete must be signed by private key. Can be read by everybody.

Supported signature scheme :
 - nacl
 - ecdsa

## Docker Install

Dependencies: `docker` & `docker-compose`


## Using the provided script

Note: modify `seed` in `docker-compose.yml` server command to change hidden service address.

### Build docker images

```bash
bash soroban.sh build
```

### Start server services

```bash
bash soroban.sh server_start
```

### Stop server services

```bash
bash soroban.sh server_stop
```

## Server services status

```bash
bash soroban.sh server_status
```

## Start the clients

Note: modify `url` in `docker-compose.yml` (`clients/python` & `clients/java`) regarding hidden service `onion` address.

```bash
bash soroban.sh clients_start
```

## Stop the clients

```bash
bash soroban.sh clients_stop
```

## Logs the clients

```bash
bash soroban.sh clients_python_logs
```

```bash
bash soroban.sh clients_java_logs
```

## Monitoring

Api entpoint for service status can be reached on `/status`

Query string `filters` can be use to filter additional information.

- `default` (`cpu,clients,keyspace`)
- `cpu`
- `clients`
- `keyspace`
- `memory`
- `stats`

Default: 

```bash
curl -s --socks5-hostname 0.0.0.0:9050 -X GET -o - http://sorzvujomsfbibm7yo3k52f3t2bl6roliijnm7qql43bcoe2kxwhbcyd.onion/status?filters=cpu,clients,keyspace
```

Wildcard: 

```bash
curl -s --socks5-hostname 0.0.0.0:9050 -X GET -o - http://sorzvujomsfbibm7yo3k52f3t2bl6roliijnm7qql43bcoe2kxwhbcyd.onion/status?filters=*
```

## Development

### Generate onion address with prefix

```bash
go run cmd/server/main.go -prefix sor
```

Output

```bash
Address: sorlnhjsp6xhb4zbqdkcr6igglar4hys3u45sofcft3ttdzqujlnutad.onion
Private Key: WAzaLtgzk5Ucd/YDkjk0PN3DiPaO0RBwVKnMOHipX3X1S7yspIRBHKweopl8wjv/EXXReFiOun5eCrZ8hUxcKg==
Seed:  169fc9f1925eec11b6a728044c9f4e6dd1a676a4f4e6f640c4100015644914e8
```

### Start soroban server with specified hostname and port

```bash
go run cmd/server/main.go --hostname=0.0.0.0 --port=4242
```

### Start soroban server with generated seed

```bash
go run cmd/server/main.go --withTor=true -seed 5baa80270886506c6b080de4e9558e2c32c50d3a7633f87d8396f5d5767e988d
```

### Export hidden service secret key

```bash
go run cmd/server/main.go -seed 5baa80270886506c6b080de4e9558e2c32c50d3a7633f87d8396f5d5767e988d -export hs_ed25519_secret_key
```

### Peer-to-Peer network

Soroban can be connect in peer-to-peer. This feature need to enable `withTor` option.

Peer descovery is done using a `p2pBootstrap` onion address to perform peer discovery.

```bash
go run cmd/server/main.go --p2pBootstrap /onion3/b6jza7z6dil564arui6gev4fmzzrppng62ixjyh66xn7fl227igs56id:1042/p2p/16Uiu2HAmKrVASuXgi7NsJZuVYu2Xqx82NkGewcfEuHKZuHq7adjB --withTor=true --seed c2d0b9870b89b10a47aa7e33fd3b51dc86eaa160d764e3b16ad3924356cc84d9
```

An optional `p2pSeed` can be used (see `prefix`) to get an well known onion address (`auto` generate a new address on startup).

When using several soroban on the same server, optional `p2pListenPort` can be use.

An optional `p2pRoom` can be use to segregate cluster on the peer-to-peer network and to not interact with other peers an another cluster.


## License

AGPL 3.0
