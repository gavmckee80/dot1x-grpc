#!/bin/bash

set -e

echo "[*] Starting bulk configuration of eth1 to eth8..."
go run examples/bulk_config.go

echo "[*] Waiting for interfaces to authenticate..."
sleep 15

echo "[*] Disconnecting interfaces..."
go run examples/bulk_disconnect.go

echo "[✔] Demo completed."
