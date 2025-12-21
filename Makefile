# Init .env file
.PHONY: init
init: install-tools install-dev-tools
	
install-tools: install-build-tools install-dev-tools
	
install-build-tools:
	go install golang.org/x/tools/cmd/goimports@latest
	go install golang.org/x/tools/cmd/stringer@latest
	go install github.com/google/wire/cmd/wire@latest
	go install go.uber.org/mock/mockgen@latest

install-dev-tools:
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
	go install golang.org/x/vuln/cmd/govulncheck@latest

.PHONY: di
di:
	wire ./cmd/obs-scheduler

.PHONY: mock
mock:
	go generate ./...

.PHONY: deps
deps:
	go mod download
	go mod tidy

.PHONY: lint
lint:
	golangci-lint run ./...
	
.PHONY: vuln-check
vuln-check:
	govulncheck ./...

.PHONY: generate
generate: di mock

START ?= 08:44
STOP ?= 10:00

start-rec:
	go run ./cmd/obs-scheduler -start $(START) -stop $(STOP)

start-rec-skip-launch:
	go run ./cmd/obs-scheduler -start $(START) -stop $(STOP) -skip-launch
