.PHONY: gifs logcopter-generate logcopter-check test test-standalone generate-xgoja build-xgoja smoke-chunking smoke-markdown smoke-sanitize smoke-extract smoke-verbs smoke check

all: gifs

VERSION=v0.1.14
GORELEASER_ARGS ?= --skip=sign --snapshot --clean
GORELEASER_TARGET ?= --single-target
XGOJA_CMD_DIR ?= cmd/goja-text
XGOJA_BINARY ?= ./dist/goja-text

TAPES=$(wildcard doc/vhs/*tape)
gifs: $(TAPES)
	for i in $(TAPES); do vhs < $$i; done

docker-lint:
	docker run --rm -v $(shell pwd):/app -w /app golangci/golangci-lint:latest golangci-lint run -v

lint:
	GOWORK=off golangci-lint run -v

lintmax:
	GOWORK=off golangci-lint run -v --max-same-issues=100

gosec:
	GOWORK=off go install github.com/securego/gosec/v2/cmd/gosec@latest
	gosec -exclude-generated -exclude=G101,G304,G301,G306 -exclude-dir=.history ./...

govulncheck:
	GOWORK=off go install golang.org/x/vuln/cmd/govulncheck@latest
	govulncheck ./...

test:
	go test ./... -count=1

test-standalone:
	GOWORK=off go test ./... -count=1

generate-xgoja:
	cd $(XGOJA_CMD_DIR) && GOWORK=off go generate

build-xgoja: generate-xgoja
	cd $(XGOJA_CMD_DIR) && GOWORK=off go build -o ../../$(XGOJA_BINARY) .

smoke-markdown: build-xgoja
	$(XGOJA_BINARY) run examples/js/markdown-demo.js

smoke-chunking: build-xgoja
	$(XGOJA_BINARY) run examples/js/chunking-demo.js

smoke-sanitize: build-xgoja
	$(XGOJA_BINARY) run examples/js/sanitize-demo.js

smoke-extract: build-xgoja
	$(XGOJA_BINARY) run examples/js/extract-demo.js

smoke-verbs: build-xgoja
	$(XGOJA_BINARY) help goja-text-chunking-user-guide >/dev/null
	$(XGOJA_BINARY) chunking pack examples/markdown/chunking-sample.md --max-units 180 --output json >/dev/null
	$(XGOJA_BINARY) help goja-text-markdown-user-guide >/dev/null
	$(XGOJA_BINARY) examples tour --output json >/dev/null
	$(XGOJA_BINARY) markdown toc examples/markdown/sample.md --output json >/dev/null
	$(XGOJA_BINARY) sanitize json examples/json/broken.json --output json >/dev/null
	$(XGOJA_BINARY) extract validate examples/text/structured-data-sample.md --output json >/dev/null

smoke: smoke-chunking smoke-markdown smoke-sanitize smoke-extract smoke-verbs

check: test test-standalone build-xgoja smoke

build: build-xgoja
	GOWORK=off go generate ./...
	GOWORK=off go build ./...

logcopter-generate:
	GOWORK=off go generate ./...

logcopter-check:
	GOWORK=off go tool logcopter-gen -area-prefix go-go-golems.goja-text -strip-prefix github.com/go-go-golems/goja-text -check ./pkg/...

goreleaser:
	GOWORK=off goreleaser release $(GORELEASER_ARGS) $(GORELEASER_TARGET)

tag-major:
	git tag $(shell svu major)

tag-minor:
	git tag $(shell svu minor)

tag-patch:
	git tag $(shell svu patch)

release:
	git push origin --tags
	GOWORK=off GOPROXY=proxy.golang.org go list -m github.com/go-go-golems/goja-text@$(shell svu current)

bump-go-go-golems:
	@deps="$$(awk '/^require[[:space:]]+github\.com\/go-go-golems\// { print $$2 } /^[[:space:]]*github\.com\/go-go-golems\// { print $$1 }' go.mod | sort -u)"; \
	if [ -z "$$deps" ]; then \
		echo "No github.com/go-go-golems dependencies in go.mod"; \
	else \
		echo "Bumping go-go-golems dependencies:"; \
		echo "$$deps"; \
		for dep in $$deps; do GOWORK=off go get "$${dep}@latest"; done; \
	fi
	GOWORK=off go mod tidy

XXX_BINARY=$(shell which XXX)
install:
	GOWORK=off go build -o ./dist/XXX ./cmd/XXX && \
		cp ./dist/XXX $(XXX_BINARY)
