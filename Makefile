APP_NAME    := pain_tz
PKG         := github.com/Alice/pain_tz
VERSION     ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT      ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE  ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

GO          := go
GOFLAGS     := -trimpath
LDFLAGS     := -s -w \
	-X $(PKG)/pkg/version.Version=$(VERSION) \
	-X $(PKG)/pkg/version.Commit=$(COMMIT) \
	-X $(PKG)/pkg/version.BuildDate=$(BUILD_DATE)

OUT_DIR     := bin
COVER_DIR   := coverage

.PHONY: all build clean test test-cover lint fmt vet install cross-build

all: fmt vet test build

build:
	@mkdir -p $(OUT_DIR)
	CGO_ENABLED=0 $(GO) build $(GOFLAGS) -ldflags="$(LDFLAGS)" -o $(OUT_DIR)/$(APP_NAME) ./cmd/$(APP_NAME)/

cross-build:
	@mkdir -p $(OUT_DIR)
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 $(GO) build $(GOFLAGS) -ldflags="$(LDFLAGS)" -o $(OUT_DIR)/$(APP_NAME)-linux-amd64 ./cmd/$(APP_NAME)/
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 $(GO) build $(GOFLAGS) -ldflags="$(LDFLAGS)" -o $(OUT_DIR)/$(APP_NAME)-linux-arm64 ./cmd/$(APP_NAME)/

clean:
	rm -rf $(OUT_DIR) $(COVER_DIR)

test:
	$(GO) test -race -count=1 -timeout=30s ./internal/... ./pkg/...

test-cover:
	@mkdir -p $(COVER_DIR)
	$(GO) test -race -count=1 -timeout=30s -coverprofile=$(COVER_DIR)/cover.out ./internal/... ./pkg/...
	$(GO) tool cover -html=$(COVER_DIR)/cover.out -o $(COVER_DIR)/cover.html

lint:
	golangci-lint run ./...

fmt:
	$(GO) fmt ./...

vet:
	$(GO) vet ./...

install: build
	sudo cp $(OUT_DIR)/$(APP_NAME) /usr/local/bin/
	sudo mkdir -p /etc/$(APP_NAME)
	sudo cp configs/$(APP_NAME).yaml /etc/$(APP_NAME)/
	sudo cp deployments/systemd/$(APP_NAME).service /etc/systemd/system/
	sudo systemctl daemon-reload
