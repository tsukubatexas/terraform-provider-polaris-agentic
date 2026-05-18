GO ?= go

.PHONY: all generate fmt test build agentic-update

all: generate fmt test build

generate:
	$(GO) run ./cmd/polaris-provider-gen -release "$${POLARIS_RELEASE:-latest}"

fmt:
	$(GO) fmt ./...

test:
	$(GO) test ./...

build:
	mkdir -p dist
	$(GO) build -o dist/terraform-provider-polaris .

agentic-update:
	scripts/agentic_loop.sh
