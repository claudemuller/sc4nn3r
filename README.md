# sc4nn3r

A port scanner written in Go.

## Requirements

- Go

## Building

```bash
go build cmd/main.go
```

## Running

| param | description | default | example |
|---|---|---|---|
| `-host` | the host to scan | | `127.1` |
| `-ports` | the port(s) or range of ports to scan | `1-1024` | `1337` |
| `-proto` | the protocol to use when scanning | `tcp` | `tcp` |

```bash
go run cmd/main.go
```

