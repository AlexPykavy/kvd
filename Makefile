.PHONY: bench bench-prof test swagger swagger-fmt

BENCH=.
SWAG := swag

bench:
	go test -run '^$$' -bench=$(BENCH) -benchtime=1000000x -benchmem -v ./... 2>&1

bench-prof:
	go test -run '^$$' -bench=$(BENCH) -benchtime=1000000x -benchmem -cpuprofile=bench.cpu.pprof -memprofile=bench.mem.pprof -v ./internal/store 2>&1

test:
	go test -race -cover -v ./... 2>&1

swagger:
	$(SWAG) fmt -d ./api/v1
	$(SWAG) init -d ./api/v1 -g docs.go -o ./docs/v1 --parseDependency --parseInternal

swagger-fmt:
	$(SWAG) fmt
