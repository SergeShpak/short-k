.PHONY: test
test:
	go test -v -coverpkg=./... -coverprofile=profile.cov ./... && \
	go tool cover -func profile.cov

.PHONY: test-clean
test-clean:
	rm -f profile.cov

GOLANGCI_LINT_CACHE?=/tmp/shortik-golangci-lint-cache

.PHONY: golangci-lint-run
golangci-lint-run: _golangci-lint-run

.PHONY: _golangci-lint-reports-mkdir
_golangci-lint-reports-mkdir:
	mkdir -p ./.golangci-lint

.PHONY: _golangci-lint-run
_golangci-lint-run: _golangci-lint-reports-mkdir
	-docker run --rm \
    -v $(shell pwd):/app \
    -v $(GOLANGCI_LINT_CACHE):/root/.cache \
    -w /app \
    golangci/golangci-lint:v1.57.2 \
        golangci-lint run \
            --config .golangci.yml \
			--out-format line-number \
	> ./.golangci-lint/report

.PHONY: golangci-lint-clean
golangci-lint-clean:
	rm -rf ./.golangci-lint

.PHONY: clean
clean: test-clean golangci-lint-clean
