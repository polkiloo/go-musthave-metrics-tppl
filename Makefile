# ========= Config =========
PLATFORM         ?= linux/amd64
GOIMAGE          ?= golang:1.24.5-bookworm

CACHE_PREFIX     ?= $(HOME)/.cache/go-$(PLATFORM)
GOMODCACHE_DIR   ?= $(CACHE_PREFIX)/mod
GOBUILDCACHE_DIR ?= $(CACHE_PREFIX)/build

.PHONY: race-docker ensure-dirs coverage

race-docker: ensure-dirs
	docker pull $(GOIMAGE)
	docker run --rm --platform=$(PLATFORM) \
		-e DEBIAN_FRONTEND=noninteractive \
		-v $(PWD):/src:delegated -w /src \
		-v $(GOMODCACHE_DIR):/go/pkg/mod \
		-v $(GOBUILDCACHE_DIR):/root/.cache/go-build \
		$(GOIMAGE) \
		bash -lc 'set -euo pipefail; \
			GO_BIN="$$(command -v go || echo /usr/local/go/bin/go)"; \
			apt-get update -qq && apt-get install -y -qq build-essential git >/dev/null; \
			CGO_ENABLED=1 "$$GO_BIN" test -race ./...'

ensure-dirs:
	@mkdir -p "$(GOMODCACHE_DIR)" "$(GOBUILDCACHE_DIR)"
	
coverage:
	@TMP=$$(mktemp); \
	if go test -coverprofile=coverage.out ./... >$$TMP; then \
		go tool cover -func=coverage.out | awk '/^total:/ {print $$3}'; \
		STATUS=0; \
	else \
		cat $$TMP; \
		STATUS=1; \
	fi; \
	RM_FILES="$$TMP coverage.out"; \
	rm -f $$RM_FILES; \
	exit $$STATUS
