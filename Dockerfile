FROM golang:1.25-alpine AS builder
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

CMD ["subconvergo"]
