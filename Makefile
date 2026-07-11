.PHONY: build test lint clean

build:
	go build -o bin/meshtastic-poi ./cmd/meshtastic-poi

test:
	go test -race ./...

lint:
	go vet ./...
	@test -z "$$(gofmt -l .)"

clean:
	rm -rf bin/

install:
	go install ./cmd/meshtastic-poi
