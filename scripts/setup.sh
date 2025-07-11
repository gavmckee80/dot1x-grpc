#!/bin/bash
set -e

echo "Setting up the Dot1x GRPC/DBus project..."
sudo apt-get update && sudo apt-get install -y protobuf-compiler dbus wpasupplicant
cd $(dirname "$0")/..
go mod init github.com/gavmckee80/dot1x-grpc
go mod tidy
make proto
make build
echo "Setup complete."
