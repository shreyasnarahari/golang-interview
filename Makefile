# golang-interview — common tasks
#
# make build          compile all packages
# make test           run the reference solution tests (green)
# make race           solution tests under the race detector
# make vet            go vet all packages
# make fmt            fail if any file is not gofmt-clean
# make exercise T=06  run module 06's exercise tests (red until you implement the stubs)
# make capstone C=rate-limiter  run a capstone's acceptance tests (red until implemented)
# make mdcheck        compile-check complete Go programs embedded in markdown
# make anki           export question banks to anki.tsv (Anki-importable deck)
# make lint           golangci-lint (if installed)
# make ci             everything CI runs

.PHONY: build test race vet fmt exercise capstone mdcheck anki lint ci

build:
	go build ./...

test:
	go test ./core-topics/*/solutions/ ./capstones/*/solution/

race:
	go test -race -count=3 ./core-topics/*/solutions/ ./capstones/*/solution/

vet:
	go vet ./...

fmt:
	@test -z "$$(gofmt -l .)" || (echo "not gofmt-clean:"; gofmt -l .; exit 1)

exercise:
	@if [ -z "$(T)" ]; then echo "usage: make exercise T=06"; exit 1; fi
	go test -v ./core-topics/$(T)-*/exercises/

capstone:
	@if [ -z "$(C)" ]; then echo "usage: make capstone C=rate-limiter"; exit 1; fi
	go test -v ./capstones/$(C)/

mdcheck:
	go run ./tools/mdcheck

anki:
	go run ./tools/ankiexport

lint:
	@command -v golangci-lint >/dev/null 2>&1 || (echo "golangci-lint not installed; skipping"; exit 0)
	golangci-lint run ./...

ci: fmt vet build test mdcheck
