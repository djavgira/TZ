APP_NAME    := tz
PKG         := github.com/djavgira/TZ
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

DOCKER_IMAGE ?= tz
DOCKER_TAG    ?= latest
DOCKERFILE    ?= Dockerfile

.PHONY: all build clean test test-cover lint fmt vet install cross-build \
        docker-build docker-run docker-push docker-clean

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

# --- Docker targets ---

docker-build:
	docker build \
		--build-arg VERSION=$(VERSION) \
		--build-arg COMMIT=$(COMMIT) \
		--build-arg BUILD_DATE=$(BUILD_DATE) \
		-t $(DOCKER_IMAGE):$(DOCKER_TAG) \
		-f $(DOCKERFILE) .

docker-build-nc:  # "no-cache" build for CI
	docker build --no-cache \
		--build-arg VERSION=$(VERSION) \
		--build-arg COMMIT=$(COMMIT) \
		--build-arg BUILD_DATE=$(BUILD_DATE) \
		-t $(DOCKER_IMAGE):$(DOCKER_TAG) \
		-f $(DOCKERFILE) .

docker-run:
	docker run --rm -it \
		--name $(APP_NAME) \
		--pid=host \
		--read-only \
		--tmpfs /tmp:size=1M,mode=1777 \
		-v /proc:/host/proc:ro \
		-v /sys:/host/sys:ro \
		-e HOST_PROC=/host/proc \
		-e HOST_SYS=/host/sys \
		-p 9100:9100 \
		$(DOCKER_IMAGE):$(DOCKER_TAG)

docker-compose-up:
	docker-compose up -d

docker-compose-down:
	docker-compose down

docker-push:
	docker push $(DOCKER_IMAGE):$(DOCKER_TAG)

docker-clean:
	docker rmi $(DOCKER_IMAGE):$(DOCKER_TAG) || true
	docker image prune -f
