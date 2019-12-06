# Soroban server in go

## Start the server

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
