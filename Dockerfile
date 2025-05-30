# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /go/src/github.com/adamfordyce11/profile-api/

COPY src/go.mod src/go.sum ./

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download

COPY src/ .

RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux \
    go build -a -mod=readonly -ldflags='-w -s -extldflags "-static"' .

# Final Stage
FROM alpine:latest

LABEL org.opencontainers.image.source="https://github.com/adamfordyce11/profile-api" \
      org.opencontainers.image.licenses="MIT" \
      org.opencontainers.image.title="profile-api" \
      org.opencontainers.image.description="Lightweight profile API service written in Go" \
      org.opencontainers.image.authors="Adam Fordyce adam@fordyce.space"


RUN apk --no-cache add ca-certificates tzdata curl openssl bash && \
    update-ca-certificates && \
    apk upgrade libssl3 libcrypto3

WORKDIR /app

# Copy the binary and necessary files from the builder stage
COPY --from=builder /go/src/github.com/adamfordyce11/profile-api/profile-api /usr/local/bin/

EXPOSE 8080

CMD ["profile-api"]
