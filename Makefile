# 单元测试
.PHONY: test
test:
	go test -race ./...

.PHONY: lint
lint:
	golangci-lint run --config .github/linters/.golangci.yml

.PHONY: fmt
fmt:
	@goimports -l -w .

.PHONY: tidy
tidy:
	@go mod tidy -v

.PHONY: check
check:
	@$(MAKE) fmt
	@$(MAKE) tidy
