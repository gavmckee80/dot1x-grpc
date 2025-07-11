# README.md

# Dot1x GRPC D-Bus Service

This service provides a gRPC API for managing 802.1X authentication on Linux Ethernet interfaces, interfacing with `wpa_supplicant` via D-Bus. It supports EAP-PEAP, EAP-TLS, EAP-TTLS, and other common methods.

---

## 🛠 Build Instructions

```bash
# Ensure Go is installed
make build
```

Or manually:
```bash
go mod tidy
go build -o bin/dot1x-server ./cmd/server
```

---

## ✅ Run Tests

```bash
make test
```

Or manually:
```bash
go test -v ./test/...
```

All unit and integration tests use `bufconn` with mocked D-Bus backend.

---

## 🚀 Run the Service

```bash
./bin/dot1x-server
```

- gRPC on `:50051`
- Prometheus metrics on `:9090/metrics`

---

## 🐳 Docker Deployment

### Dockerfile
Create a file named `Dockerfile`:
```Dockerfile
FROM golang:1.21 as builder
WORKDIR /app
COPY . .
RUN go build -o dot1x-server ./cmd/server

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y dbus wpasupplicant && rm -rf /var/lib/apt/lists/*
COPY --from=builder /app/dot1x-server /usr/local/bin/
ENTRYPOINT ["/usr/local/bin/dot1x-server"]
```

### Build and Run
```bash
docker build -t dot1x-server .
docker run --rm --net=host --privileged dot1x-server
```

> Note: `--net=host` and `--privileged` are required to access D-Bus and network interfaces inside the container.

---

## 📡 gRPC Interface

Use tools like `grpcurl` or generate stubs with:
```bash
protoc --go_out=. --go-grpc_out=. proto/ether8021x.proto
```

Then build a client or use `grpcurl`:
```bash
grpcurl -plaintext localhost:50051 list
```

---

## 📁 Project Structure

- `cmd/server/` – gRPC server entrypoint
- `internal/core/` – Business logic and validation
- `internal/dbus/` – D-Bus abstraction to wpa_supplicant
- `internal/grpc/` – gRPC service implementation
- `proto/` – gRPC protobuf definitions
- `test/` – Unit tests and mocks
- `scripts/setup.sh` – Project setup script

---

## ✨ Features
- Full test coverage with mocks
- Secure TLS credential handling
- Prometheus metrics integration
- Graceful shutdown
- Concurrent request safety
- gRPC streaming for live interface events

---

## 🧪 Example
Run a simple PEAP authentication:
```grpc
ConfigureInterface {
  interface: "eth0"
  eap_type: EAP_PEAP
  identity: "alice"
  password: "password"
  phase2_auth: "mschapv2"
}
```

---

## 📜 License




## Notes:

```
docker build -t dot1x-server .
docker run --rm --net=host --privileged dot1x-server
```

```
docker build -f dot1xctl.Dockerfile -t dot1xctl .
docker run --rm --net=host --privileged dot1xctl -status
```


Dbus service install note

```
sudo cp dot1x.service /etc/systemd/system/
sudo systemctl daemon-reexec
sudo systemctl enable --now dot1x.service
```