.PHONY: check complexity fuzz-short lint race test tools

check: test lint complexity

test:
	go test ./...

race:
	go test -race ./...

lint: tools
	./scripts/lint.sh

complexity: tools
	./scripts/complexity.sh

fuzz-short:
	go test ./internal/cmd -run '^$$' -fuzz FuzzSplitCommandLine -fuzztime=2s
	go test ./internal/cmd -run '^$$' -fuzz FuzzParseBitbucketEntityURL -fuzztime=2s

tools:
	./scripts/install-dev-tools.sh
