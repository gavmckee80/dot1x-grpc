# Build the CLI tool
FROM golang:1.21 as builder

WORKDIR /cli
COPY . .
RUN go build -o dot1xctl ./cmd/dot1xctl

# Minimal runtime
FROM debian:bookworm-slim

COPY --from=builder /cli/dot1xctl /usr/local/bin/dot1xctl
ENTRYPOINT ["/usr/local/bin/dot1xctl"]
