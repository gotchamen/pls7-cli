.PHONY: build test test-v test-race test-cover run run-dev run-nlh clean lint vet fmt check

# Build
build:
	go build -v ./...

# Run
run:
	go run main.go

run-dev:
	go run main.go --dev

run-nlh:
	go run main.go -r nlh

# Test
test:
	go test ./...

test-v:
	go test -v ./...

test-race:
	go test -race ./...

test-cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out
	@rm -f coverage.out

test-cover-html:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Open coverage.html in your browser"
	@rm -f coverage.out

# Code quality
vet:
	go vet ./...

fmt:
	gofmt -w .

fmt-check:
	@test -z "$$(gofmt -l .)" || (echo "Files need formatting:" && gofmt -l . && exit 1)

lint: vet fmt-check

# CI-equivalent check (matches .github/workflows/go.yml)
check: build test-v

# Clean
clean:
	go clean
	rm -f coverage.out coverage.html
