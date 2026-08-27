.PHONY: bench bench-prof k6 test swagger swagger-fmt

BENCH=.

ROOT_DIR := $(dir $(realpath $(lastword $(MAKEFILE_LIST))))
SWAG := swag

bench:
	go test -run '^$$' -bench=$(BENCH) -benchtime=1000000x -benchmem -v ./... 2>&1

bench-prof:
	go test -run '^$$' -bench=$(BENCH) -benchtime=1000000x -benchmem -cpuprofile=bench.cpu.pprof -memprofile=bench.mem.pprof -v ./internal/store 2>&1

test:
	go test -race -cover -v ./... 2>&1

k6:
	docker run --rm \
		--name kvd-k6 \
		-e BASE_URL="http://kvd:8080" \
		-v $(ROOT_DIR)/k6:/opt/k6:ro \
		--network kvd_default \
		--cpus="4" \
		grafana/k6:2.2.0 \
		run \
		/opt/k6/00-concurrent-health-check.js

swagger:
	$(SWAG) fmt -d ./api/v1
	$(SWAG) init -d ./api/v1 -g docs.go -o ./docs/v1 --parseDependency --parseInternal

swagger-fmt:
	$(SWAG) fmt
