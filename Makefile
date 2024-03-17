.PHONY: lint
lint:
	golangci-lint run

.PHONY: lint-dev
lint-dev:
	golangci-lint run --tests=false

.PHONY: generate
generate:
	go generate ./...

.PHONY: test
test:
	AUTO_CREATE_CI_DB=true go test ./...