.PHONY: benchmark
benchmark:
	go test -bench . ./...

.PHONY: install
install:
	go install ./cmd/claudelaunch

.PHONY: cover
cover:
	go tool cover -html cover.out

.PHONY: fmt
fmt:
	goimports -w -local `head -n 1 go.mod | sed 's/^module //'` .

.PHONY: lint
lint:
	golangci-lint run

.PHONY: restart
restart: install
	-tmux kill-session -t claudelaunch-server
	tmux new-session -d -s claudelaunch-server ~/Developer/go/bin/claudelaunch

.PHONY: run
run:
	go run ./cmd/claudelaunch

.PHONY: test
test:
	go test -coverprofile cover.out -shuffle on ./...
