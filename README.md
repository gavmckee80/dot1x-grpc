# README.md

# Dot1x GRPC D-Bus Service

This service provides a gRPC API for managing 802.1X authentication on Linux Ethernet interfaces, interfacing with `wpa_supplicant` via D-Bus. It supports EAP-PEAP, EAP-TLS, EAP-TTLS, and other common methods.

## ✨ Features

- **gRPC API** for 802.1X authentication management
- **D-Bus Integration** with wpa_supplicant
- **Multiple EAP Methods** (PEAP, TLS, TTLS, FAST)
- **gRPC Reflection** for service discovery and testing
- **Comprehensive Testing** with mocked D-Bus backend
- **Secure TLS Credential Handling**
- **Graceful Shutdown** and resource cleanup
- **Concurrent Request Safety**
- **Real-time Status Streaming**

---

## 🛠 Build Instructions

```bash
# Build all binaries
make build
```

Or manually:
```bash
go mod tidy
go build -o bin/dot1x-server ./cmd/server
go build -o bin/dot1x-cli ./cmd/cli
go build -o bin/test-server ./cmd/test-server
```

---

## 🚀 Run the Service

### Production Server
```bash
./bin/dot1x-server
```

- gRPC on `:50051`
- Prometheus metrics on `:9090/metrics`
- Requires D-Bus system connection

### Test Server (No D-Bus Required)
For development and testing without D-Bus:
```bash
./bin/test-server
```

- Uses mock D-Bus client
- gRPC reflection enabled
- Perfect for API exploration and testing

### CLI Client
```bash
./bin/dot1x-cli -interface eth0 -eap PEAP -identity user -password pass
```

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

# Stream status updates
grpcurl -plaintext -d '{"interface": "eth0"}' localhost:50051 ether8021x.Dot1xManager/StreamStatus
```

### Manual Stub Generation

For production clients, generate stubs with:
```bash
protoc --go_out=. --go-grpc_out=. proto/ether8021x.proto
```

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

## ✅ Testing

### Run All Tests
```bash
make test
```

Or manually:
```bash
go test -v ./test/...
```

### Test Coverage
```bash
go test -cover ./...
```

All unit and integration tests use `bufconn` with mocked D-Bus backend.

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

## 📁 Project Structure

```
dot1x-grpc/
├── cmd/
│   ├── server/          # Production gRPC server
│   ├── cli/            # Command-line client
│   └── test-server/    # Test server with mock D-Bus
├── internal/
│   ├── core/           # Business logic and validation
│   ├── dbus/           # D-Bus abstraction to wpa_supplicant
│   └── grpc/           # gRPC service implementation
├── proto/              # gRPC protobuf definitions
├── test/               # Unit tests and mocks
├── examples/           # Usage examples
└── scripts/            # Setup and utility scripts
```

---

## 🐳 Docker Deployment

### Production Server
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

## 🧪 Usage Examples

### Configure PEAP Authentication
```bash
grpcurl -plaintext -d '{
  "interface": "eth0",
  "eap_type": "EAP_PEAP",
  "identity": "alice",
  "password": "password",
  "phase2_auth": "mschapv2"
}' localhost:50051 ether8021x.Dot1xManager/ConfigureInterface
```

### Configure TLS Authentication
```bash
grpcurl -plaintext -d '{
  "interface": "eth0",
  "eap_type": "EAP_TLS",
  "identity": "cert-user",
  "ca_cert": "base64-encoded-ca-cert",
  "client_cert": "base64-encoded-client-cert",
  "private_key": "base64-encoded-private-key"
}' localhost:50051 ether8021x.Dot1xManager/ConfigureInterface
```

### Get Interface Status
```bash
grpcurl -plaintext -d '{"interface": "eth0"}' localhost:50051 ether8021x.Dot1xManager/GetStatus
```

### Disconnect Interface
```bash
grpcurl -plaintext -d '{"interface": "eth0"}' localhost:50051 ether8021x.Dot1xManager/Disconnect
```

---

## 🔧 Troubleshooting

### D-Bus Connection Issues
If you get D-Bus connection errors on macOS:
```bash
# Use the test server instead
./bin/test-server
```

### gRPC Reflection Not Working
Ensure the server is running and try:
```bash
# Check if server is listening
netstat -an | grep 50051

# Test reflection
grpcurl -plaintext localhost:50051 list
```

### Build Issues
```bash
# Clean and rebuild
make clean
make build

# Regenerate protobuf stubs
./proto/generate.sh
```

---

## 📜 License

[Add your license information here]

---

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Ensure all tests pass
6. Submit a pull request

---

## 📝 Notes

### Systemd Service Installation
```bash
sudo cp dot1x.service /etc/systemd/system/
sudo systemctl daemon-reexec
sudo systemctl enable --now dot1x.service
```

### Docker CLI Client
```bash
docker build -f dot1xctl.Dockerfile -t dot1xctl .
docker run --rm --net=host --privileged dot1xctl -status
```
