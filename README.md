# Soroban server in go

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

### Start soroban server with generated seed

```bash
go run cmd/server/main.go -seed 5baa80270886506c6b080de4e9558e2c32c50d3a7633f87d8396f5d5767e988d
```

### Export hidden service secret key

```bash
go run cmd/server/main.go -seed 5baa80270886506c6b080de4e9558e2c32c50d3a7633f87d8396f5d5767e988d -export hs_ed25519_secret_key
```

## License

AGPL 3.0
