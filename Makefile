.PHONY: check complexity fuzz-short lint race stability test test-repeat test-shuffle tools

check: test lint complexity

test:
	go test ./...

test-shuffle:
	go test -shuffle=on ./...

test-repeat:
	go test -count=2 ./...

stability: test-shuffle test-repeat

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
