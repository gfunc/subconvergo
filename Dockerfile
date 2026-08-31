# Pin the builder to an explicit patched Go patch release (not a floating
# minor tag): go1.26.7 includes the stdlib security fixes flagged by
# govulncheck (net/url, crypto/tls, net/http, encoding/asn1; see
# .scratch/security-hardening/issues/07). Keep in sync with the go directive
# in go.mod.
FROM golang:1.26.7-alpine AS builder
# proxy.golang.org is unreachable from the networks where this image is
# usually built (the apk mirror below is already tuna for the same reason).
# goproxy.cn proxies both module zips and the sum.golang.org checksum
# database, so go.sum verification (default GOSUMDB) still applies.
ENV GOPROXY=https://goproxy.cn

WORKDIR /app

# Install dependencies
RUN sed -i 's#https\?://dl-cdn.alpinelinux.org/alpine#https://mirrors.tuna.tsinghua.edu.cn/alpine#g' /etc/apk/repositories
RUN apk add --no-cache git gcc musl-dev curl

# Create directories that might be needed with proper permissions
RUN mkdir -p /go/pkg /go/bin && chmod -R 777 /go
RUN mkdir /.cache && chmod -R 777 /.cache

# Copy go mod files
COPY go.mod go.sum* ./
RUN go mod download

# Copy source code
COPY . .

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o subconvergo main.go

# Final stage
FROM alpine:3.21


RUN sed -i 's#https\?://dl-cdn.alpinelinux.org/alpine#https://mirrors.tuna.tsinghua.edu.cn/alpine#g' /etc/apk/repositories
RUN apk --no-cache add ca-certificates tzdata
# Copy binary
COPY --from=builder /app/subconvergo /usr/bin/subconvergo

# Copy base configuration from parent directory
COPY base /base

# Set environment
ENV TZ=UTC
RUN ln -sf /usr/share/zoneinfo/$TZ /etc/localtime && \
    echo $TZ > /etc/timezone

# Create non-root user
RUN addgroup -g 1000 subconvergo && \
    adduser -D -u 1000 -G subconvergo subconvergo && \
    chown -R subconvergo:subconvergo /base

WORKDIR /base

USER subconvergo
EXPOSE 25500

ENTRYPOINT ["subconvergo"]
