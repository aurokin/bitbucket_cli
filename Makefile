.PHONY: check complexity lint test tools

check: test lint complexity

test:
	go test ./...

lint: tools
	./scripts/lint.sh

complexity: tools
	./scripts/complexity.sh

tools:
	./scripts/install-dev-tools.sh
