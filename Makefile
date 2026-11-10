.PHONY: all clean test docker docker-indexserver docker-webserver docker-push

# === Build binaries ===
all: indexserver webserver generator partitioner

webserver:
	go build -o bin/webserver ./cmd/webserver
indexserver:
	go build -o bin/indexserver ./cmd/indexserver

generator:
	go build -o bin/generator ./tools/generator

partitioner:
	go build -o bin/partitioner ./tools/partitioner

# === Test ===

test:
	go test ./...

# === Clean ===

clean:
	rm -f bin/indexserver bin/webserver bin/partitioner
