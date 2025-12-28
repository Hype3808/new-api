# syntax=docker/dockerfile:1.4

FROM --platform=$BUILDPLATFORM oven/bun:latest AS web-builder

WORKDIR /build
COPY ./web/package.json ./
RUN bun install
COPY ./web .
COPY ./VERSION .
RUN DISABLE_ESLINT_PLUGIN='true' VITE_REACT_APP_VERSION=$(cat VERSION) bun run build

FROM --platform=$BUILDPLATFORM golang:alpine AS builder2
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOPROXY=https://proxy.golang.org,direct

ARG TARGETOS
ARG TARGETARCH
ENV GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64}

WORKDIR /build

# Copy only go.mod and go.sum first to leverage Docker cache
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download && go mod verify

# Copy source code after dependencies are cached
COPY . .

# Copy web dist from builder
COPY --from=web-builder /build/dist ./web/dist

# Build with optimizations
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -trimpath -ldflags "-s -w -X 'github.com/QuantumNous/new-api/common.Version=$(cat VERSION)'" -o new-api

# Install certs and tzdata on build platform to avoid emulation during target builds
FROM --platform=$BUILDPLATFORM alpine:latest AS deps
RUN apk add --no-cache ca-certificates tzdata

FROM alpine:latest

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder2 /build/new-api /
EXPOSE 3000
WORKDIR /data
ENTRYPOINT ["/new-api"]
