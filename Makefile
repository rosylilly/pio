GO_FILES:=$(shell find . -type f -name '*.go')

.PHONY: test
test: tmp/cover.html

tmp:
	@mkdir -p ./tmp

tmp/cover.out: tmp $(GO_FILES)
	@go test -timeout 5s -coverprofile ./tmp/cover.out -v ./...

tmp/cover.html: tmp/cover.out
	@go tool cover -html=./tmp/cover.out -o ./tmp/cover.html
