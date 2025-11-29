# ========= Config =========
PLATFORM         ?= linux/amd64
GOIMAGE          ?= golang:1.24.5-bookworm

PROFILE_DIR             ?= profiles
NETWORK_BENCHTIME       ?= 2x
COLLECTOR_BENCHTIME     ?= 25000x
STORAGE_BENCHTIME       ?= 72897x
PROFILE_BENCH_COUNT     ?= 1

HEY_URL                 ?= http://localhost:8080/updates
HEY_PAYLOAD             ?= testdata/network_batch.json
HEY_REQUESTS            ?= 10000
HEY_CONCURRENCY         ?= 100
SKIP_HEY                ?= 0

CACHE_PREFIX     ?= $(HOME)/.cache/go-$(PLATFORM)
GOMODCACHE_DIR   ?= $(CACHE_PREFIX)/mod
GOBUILDCACHE_DIR ?= $(CACHE_PREFIX)/build

.PHONY: race-docker ensure-dirs coverage ensure-profile-dir \
        profile-network profile-collector profile-storage

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

ensure-profile-dir:
	@mkdir -p "$(PROFILE_DIR)"

profile-network: ensure-profile-dir
	@echo "Generating network handler heap profile into $(PROFILE_DIR)/network.pprof";
	@set -eu; \
		TMP_PROFILE=$$(mktemp "$(PROFILE_DIR)/network.pprof.XXXXXX"); \
		TMP_HEY=$$(mktemp "$(PROFILE_DIR)/network_hey.txt.XXXXXX"); \
		cleanup() { rm -f "$$TMP_PROFILE" "$$TMP_HEY"; }; \
		trap 'cleanup' INT TERM EXIT; \
		GOFLAGS='' go test -run=^$$ -bench=BenchmarkGinHandlerUpdatesJSONNetwork -benchmem -count=$(PROFILE_BENCH_COUNT) \
				-benchtime=$(NETWORK_BENCHTIME) -memprofile="$$TMP_PROFILE" ./internal/handler; \
		if [ "$(SKIP_HEY)" != "1" ]; then \
			command -v hey >/dev/null || { echo "hey is required. Install via 'brew install hey'" >&2; exit 1; }; \
			if [ ! -f "$(HEY_PAYLOAD)" ]; then \
				echo "payload file $(HEY_PAYLOAD) not found" >&2; \
				exit 1; \
			fi; \
			echo "Running hey load test against $(HEY_URL)"; \
			if ! hey -n $(HEY_REQUESTS) -c $(HEY_CONCURRENCY) -m POST -T application/json \
					-d @$(HEY_PAYLOAD) $(HEY_URL) >"$$TMP_HEY"; then \
				echo "hey failed to reach $(HEY_URL). Ensure the server is running or rerun with SKIP_HEY=1" >&2; \
				exit 1; \
			fi; \
			cat "$$TMP_HEY"; \
		fi; \
		mv "$$TMP_PROFILE" "$(PROFILE_DIR)/network.pprof"; \
		if [ "$(SKIP_HEY)" != "1" ]; then \
			mv "$$TMP_HEY" "$(PROFILE_DIR)/network_hey.txt"; \
		else \
			rm -f "$$TMP_HEY"; \
		fi; \
		trap - INT TERM EXIT; \
		cleanup() { :; }; \
		cleanup

profile-collector: ensure-profile-dir
	@echo "Generating collector heap profile into $(PROFILE_DIR)/collector.pprof";
	@set -eu; \
		TMP_PROFILE=$$(mktemp "$(PROFILE_DIR)/collector.pprof.XXXXXX"); \
		cleanup() { rm -f "$$TMP_PROFILE"; }; \
		trap 'cleanup' INT TERM EXIT; \
		GOFLAGS='' go test -run=^$$ -bench=BenchmarkCollectorCollect -benchmem -count=$(PROFILE_BENCH_COUNT) \
				-benchtime=$(COLLECTOR_BENCHTIME) -memprofile="$$TMP_PROFILE" ./internal/collector; \
		mv "$$TMP_PROFILE" "$(PROFILE_DIR)/collector.pprof"; \
		trap - INT TERM EXIT; \
		cleanup() { :; }; \
		cleanup

profile-storage: ensure-profile-dir
	@echo "Generating storage heap profile into $(PROFILE_DIR)/storage.pprof";
	@set -eu; \
		TMP_PROFILE=$$(mktemp "$(PROFILE_DIR)/storage.pprof.XXXXXX"); \
		cleanup() { rm -f "$$TMP_PROFILE"; }; \
		trap 'cleanup' INT TERM EXIT; \
		GOFLAGS='' go test -run=^$$ -bench=BenchmarkMemStorageSnapshot -benchmem -count=$(PROFILE_BENCH_COUNT) \
				-benchtime=$(STORAGE_BENCHTIME) -memprofile="$$TMP_PROFILE" ./internal/storage; \
		mv "$$TMP_PROFILE" "$(PROFILE_DIR)/storage.pprof"; \
		trap - INT TERM EXIT; \
		cleanup() { :; }; \
		cleanup