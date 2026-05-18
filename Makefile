BINARY   := asaguard
CMD      := ./cmd/asaguard
VERSION  := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS  := -ldflags "-X main.version=$(VERSION)"

AI_GUARDRAILS_SIEM_ENDPOINT ?= https://wnniwdexyj.execute-api.eu-west-2.amazonaws.com/

.PHONY: build test install lint clean siem-check

build:
	go build $(LDFLAGS) -o $(BINARY) $(CMD)

test:
	go test ./...

install:
	go install $(LDFLAGS) $(CMD)

lint:
	go vet ./...

siem-check: build
	AI_GUARDRAILS_SIEM_ENDPOINT=$(AI_GUARDRAILS_SIEM_ENDPOINT) ./$(BINARY) check

clean:
	rm -f $(BINARY)
