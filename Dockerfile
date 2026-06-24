# Stage 1: Build
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /src

# Dependency cache layer
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" -o /tz ./cmd/tz/

# Stage 2: Runtime
FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata && \
    adduser -D -u 65532 -s /sbin/nologin tz

# Copy static binary
COPY --from=builder /tz /usr/local/bin/tz

# Container-optimized default config
# Override via bind mount or TZ_* environment variables
COPY configs/tz.container.yaml /etc/tz/tz.yaml

USER tz
EXPOSE 9100

HEALTHCHECK --interval=15s --timeout=3s --start-period=10s --retries=3 \
    CMD wget -qO- http://localhost:9100/health || exit 1

ENTRYPOINT ["tz", "serve", "--config", "/etc/tz/tz.yaml"]
