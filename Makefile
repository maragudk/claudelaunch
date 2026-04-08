.PHONY: benchmark
benchmark:
	go test -bench . ./...

.PHONY: build
build:
	go build -o claudelaunch ./cmd/app

.PHONY: cover
cover:
	go tool cover -html cover.out

.PHONY: fmt
fmt:
	goimports -w -local `head -n 1 go.mod | sed 's/^module //'` .

.PHONY: lint
lint:
	golangci-lint run

.PHONY: run
run:
	go run ./cmd/app

.PHONY: test
test:
	go test -coverprofile cover.out -shuffle on ./...
