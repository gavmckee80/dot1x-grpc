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

### Production Server
```bash
./bin/dot1x-server
```

- gRPC on `:50051`
- Prometheus metrics on `:9090/metrics`

### Test Server (No D-Bus Required)
For development and testing without D-Bus:
```bash
./bin/test-server
```

- Uses mock D-Bus client
- gRPC reflection enabled
- Perfect for API exploration and testing

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

### Service Discovery with Reflection

The server supports gRPC reflection, allowing you to explore the API without the `.proto` files:

```bash
# List available services
grpcurl -plaintext localhost:50051 list

# Describe a service
grpcurl -plaintext localhost:50051 describe ether8021x.Dot1xManager

# Describe message types
grpcurl -plaintext localhost:50051 describe ether8021x.Dot1xConfigRequest

# Call a method
grpcurl -plaintext -d '{"interface": "eth0", "eap_type": "EAP_PEAP", "identity": "user", "password": "pass"}' localhost:50051 ether8021x.Dot1xManager/ConfigureInterface
```

### Manual Stub Generation

For production clients, generate stubs with:
```bash
protoc --go_out=. --go-grpc_out=. proto/ether8021x.proto
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

## 🧬 Generate Go Protobuf Stubs

To generate the Go gRPC and protobuf stubs from the proto definition, you need to have `protoc` and the Go plugins installed:

### 1. Install prerequisites (if not already installed):

```bash
# Install protoc (if not already installed)
brew install protobuf

# Install Go plugins for protoc (if not already installed)
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

Make sure your `$GOPATH/bin` is in your `$PATH` so `protoc` can find the plugins.

### 2. Generate the stubs:

```bash
cd proto
./generate.sh
```

This will generate/update the Go files in the correct locations for your project.

---

## 📚 Documentation

### Local Development
Serve documentation locally:
```bash
# Install godoc if not already installed
go install golang.org/x/tools/cmd/godoc@latest

# Serve documentation on localhost:6060
godoc -http=:6060
```

Then visit: **http://localhost:6060/pkg/github.com/gavmckee80/dot1x-grpc/**

### Generate Static Documentation
```bash
mkdir -p docs
godoc -url=/pkg/github.com/gavmckee80/dot1x-grpc/ > docs/index.html
```

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