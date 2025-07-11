#!/bin/bash

set -e

PROTO_DIR="$(dirname "$0")"
OUT_DIR="${PROTO_DIR}/.."

echo "[*] Generating Go protobuf stubs..."
protoc \
  --go_out="$OUT_DIR" \
  --go-grpc_out="$OUT_DIR" \
  --go_opt=paths=source_relative \
  --go-grpc_opt=paths=source_relative \
  "${PROTO_DIR}/ether8021x.proto"

echo "[✔] Done."
