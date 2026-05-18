BINARY   := asaguard
CMD      := ./cmd/asaguard
VERSION  := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS  := -ldflags "-X main.version=$(VERSION)"

.PHONY: build test install lint clean

build:
	go build $(LDFLAGS) -o $(BINARY) $(CMD)

test:
	go test ./...

install:
	go install $(LDFLAGS) $(CMD)

lint:
	go vet ./...

clean:
	rm -f $(BINARY)
