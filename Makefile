fmt:
	go fmt ./...
.PHONY: fmt

lint:
	golangci-lint run -E goimports -E godot --timeout 10m
.PHONY: lint

build:
	go build -o ./vaxwish ./cmd
.PHONY: build
