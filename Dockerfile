# Stage 1: Build
FROM golang:1.21 as builder

WORKDIR /app
COPY . .
RUN go build -o dot1x-server ./cmd/server

# Stage 2: Runtime
FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y dbus wpasupplicant && rm -rf /var/lib/apt/lists/*
COPY --from=builder /app/dot1x-server /usr/local/bin/dot1x-server

ENTRYPOINT ["/usr/local/bin/dot1x-server"]
