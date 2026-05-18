GO ?= go
SPEC_CACHE_DIR ?= specs

.PHONY: all generate fmt test build agentic-update

all: generate fmt test build

generate:
	$(GO) run ./cmd/polaris-provider-gen -release "$${POLARIS_RELEASE:-latest}" -spec-cache-dir "$(SPEC_CACHE_DIR)"

fmt:
	$(GO) fmt ./...

test:
	$(GO) test ./...

build:
	mkdir -p dist
	$(GO) build -buildvcs=false -o dist/terraform-provider-polaris .

agentic-update:
	scripts/agentic_loop.sh
