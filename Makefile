SRC := go.sum $(shell git ls-files -cmo --exclude-standard -- "*.go")
TESTABLE := ./...

# let's not build this
# use the lint action in the pr action, or a local installed version
bin/golangci-lint: GOARCH =
bin/golangci-lint: GOOS =
bin/golangci-lint: go.sum
	@go get github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2
	@go build -o $@ github.com/golangci/golangci-lint/v2/cmd/golangci-lint


lint: CGO_ENABLED = 1
lint: GOARCH =
lint: GOOS =
lint: bin/golangci-lint $(SRC)
	$< run

tidy:
	go mod tidy

dirty: tidy
	git status --porcelain
	@[ -z "$$(git status --porcelain)" ]

#vet: GOARCH =
#vet: GOOS =
#vet: CGO_ENABLED =
#vet: bin/go-acc $(SRC)
#	$< --covermode=atomic $(TESTABLE) -- -race -v