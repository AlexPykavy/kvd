.PHONY: bench test swagger swagger-fmt

SWAG := swag

bench:
	go test -run '^$$' -bench=. -benchtime=10000000x -benchmem -v ./... 2>&1

test:
	go test -race -cover -v ./... 2>&1

swagger:
	$(SWAG) fmt -d ./api/v1
	$(SWAG) init -d ./api/v1 -g docs.go -o ./docs/v1 --parseDependency --parseInternal

swagger-fmt:
	$(SWAG) fmt
