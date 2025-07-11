BINARY_NAME=dot1x-server
CLI_NAME=dot1x-cli

all: build

build:
	go build -o bin/$(BINARY_NAME) ./cmd/server
	go build -o bin/$(CLI_NAME) ./cmd/cli

proto:
	protoc \
	  --go_out=. \
	  --go-grpc_out=. \
	  --go_opt=paths=source_relative \
	  --go-grpc_opt=paths=source_relative \
	  proto/ether8021x.proto

run:
	./bin/$(BINARY_NAME)

cli:
	./bin/$(CLI_NAME)

test:
	go test -v ./test/...

demo:
	bash run_bulk_demo.sh

docker-build:
	docker build -t dot1x-server .

docker-run:
	docker run --rm --net=host --privileged dot1x-server

clean:
	rm -rf bin

.PHONY: all build run test proto docker-build docker-run clean demo cli
