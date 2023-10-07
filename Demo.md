# P2P Demo

## Soroban default room (samourai)

- Start 3 soroban servers (4242, 4243, 4244)

```
# 4242
go run ./cmd/server --p2pBootstrap=/ip4/164.68.108.59/tcp/1042/p2p/16Uiu2HAmLh9dzKen97hWAkEoRawFVQq2j1Lba56tNeoDstpzUZS4 --p2pRoom=samourai --port=4242

# 4243
go run ./cmd/server --p2pBootstrap=/ip4/164.68.108.59/tcp/1042/p2p/16Uiu2HAmLh9dzKen97hWAkEoRawFVQq2j1Lba56tNeoDstpzUZS4 --p2pRoom=samourai --port=4243

# 4244
go run ./cmd/server --p2pBootstrap=/ip4/164.68.108.59/tcp/1042/p2p/16Uiu2HAmLh9dzKen97hWAkEoRawFVQq2j1Lba56tNeoDstpzUZS4 --p2pRoom=samourai --port=4244
```

- List from first soroban server (4242)
```
# 4242
curl -s -X POST  -H 'Content-Type: application/json' -d '{ "jsonrpc": "2.0", "id": 42, "method":"directory.List", "params": [{ "Name": "foo"}] }' http://localhost:4242/rpc | jq .
```


- Add directory entry to sorobans server (4242)

```
# 4242
curl -X POST  -H 'Content-Type: application/json' -d '{ "jsonrpc": "2.0", "id": 42, "method":"directory.Add", "params": [{ "Name": "foo", "Entry": "foo_42", "Mode": "short"}] }' http://localhost:4242/rpc
# 4243
curl -X POST  -H 'Content-Type: application/json' -d '{ "jsonrpc": "2.0", "id": 42, "method":"directory.Add", "params": [{ "Name": "foo", "Entry": "foo_43", "Mode": "short"}] }' http://localhost:4243/rpc
# 4244
curl -X POST  -H 'Content-Type: application/json' -d '{ "jsonrpc": "2.0", "id": 42, "method":"directory.Add", "params": [{ "Name": "foo", "Entry": "foo_44", "Mode": "short"}] }' http://localhost:4244/rpc

```

### List from all soroban servers

```
# 4242
curl -s -X POST  -H 'Content-Type: application/json' -d '{ "jsonrpc": "2.0", "id": 42, "method":"directory.List", "params": [{ "Name": "foo"}] }' http://localhost:4242/rpc | jq .

# 4243
curl -s -X POST  -H 'Content-Type: application/json' -d '{ "jsonrpc": "2.0", "id": 42, "method":"directory.List", "params": [{ "Name": "foo"}] }' http://localhost:4243/rpc | jq .

# 4244
curl -s -X POST  -H 'Content-Type: application/json' -d '{ "jsonrpc": "2.0", "id": 42, "method":"directory.List", "params": [{ "Name": "foo"}] }' http://localhost:4244/rpc | jq .
```


## Soroban private room (samourai-wallet)

Start 3 soroban private servers (4201, 4202, 4203)

```
# 4201
go run ./cmd/server --p2pBootstrap=/ip4/164.68.108.59/tcp/1042/p2p/16Uiu2HAmLh9dzKen97hWAkEoRawFVQq2j1Lba56tNeoDstpzUZS4 --p2pRoom=samourai-wallet --port=4201

# 4202
go run ./cmd/server --p2pBootstrap=/ip4/164.68.108.59/tcp/1042/p2p/16Uiu2HAmLh9dzKen97hWAkEoRawFVQq2j1Lba56tNeoDstpzUZS4 --p2pRoom=samourai-wallet --port=4202

4203
go run ./cmd/server --p2pBootstrap=/ip4/164.68.108.59/tcp/1042/p2p/16Uiu2HAmLh9dzKen97hWAkEoRawFVQq2j1Lba56tNeoDstpzUZS4 --p2pRoom=samourai-wallet --port=4203
```


- Add directory entries to soroban private server (4201)

```
# 4201
curl -X POST  -H 'Content-Type: application/json' -d '{ "jsonrpc": "2.0", "id": 42, "method":"directory.Add", "params": [{ "Name": "foo", "Entry": "foo_01", "Mode": "short"}] }' http://localhost:4201/rpc
# 4202
curl -X POST  -H 'Content-Type: application/json' -d '{ "jsonrpc": "2.0", "id": 42, "method":"directory.Add", "params": [{ "Name": "foo", "Entry": "foo_02", "Mode": "short"}] }' http://localhost:4202/rpc
# 4203
curl -X POST  -H 'Content-Type: application/json' -d '{ "jsonrpc": "2.0", "id": 42, "method":"directory.Add", "params": [{ "Name": "foo", "Entry": "foo_03", "Mode": "short"}] }' http://localhost:4203/rpc
```

- List from all soroban private servers

```
# 4201
curl -s -X POST  -H 'Content-Type: application/json' -d '{ "jsonrpc": "2.0", "id": 42, "method":"directory.List", "params": [{ "Name": "foo"}] }' http://localhost:4201/rpc | jq .

# 4202
curl -s -X POST  -H 'Content-Type: application/json' -d '{ "jsonrpc": "2.0", "id": 42, "method":"directory.List", "params": [{ "Name": "foo"}] }' http://localhost:4202/rpc | jq .

# 4203
curl -s -X POST  -H 'Content-Type: application/json' -d '{ "jsonrpc": "2.0", "id": 42, "method":"directory.List", "params": [{ "Name": "foo"}] }' http://localhost:4203/rpc | jq .
```

- List from default soroban server

```
# 4242
curl -s -X POST  -H 'Content-Type: application/json' -d '{ "jsonrpc": "2.0", "id": 42, "method":"directory.List", "params": [{ "Name": "foo"}] }' http://localhost:4242/rpc | jq .
```