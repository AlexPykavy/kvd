.PHONY: swagger swagger-fmt

SWAG := swag

swagger:
	$(SWAG) fmt -d ./api/v1
	$(SWAG) init -d ./api/v1 -g docs.go -o ./docs/v1 --parseDependency --parseInternal

swagger-fmt:
	$(SWAG) fmt
