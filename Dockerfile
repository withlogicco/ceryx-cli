# syntax=docker/dockerfile:1

# Builds and tests run here so the host needs only Docker. The build stage runs on the
# host platform and cross-compiles for --platform (e.g. darwin/arm64).
FROM --platform=$BUILDPLATFORM golang:1.27-alpine AS base
WORKDIR /src
ENV CGO_ENABLED=0
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download
COPY . .

FROM base AS test
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build \
    go vet ./... && go test ./...

FROM base AS build
ARG TARGETOS TARGETARCH VERSION=dev
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build \
    GOOS=$TARGETOS GOARCH=$TARGETARCH go build -ldflags "-s -w -X main.version=$VERSION" -o /out/ceryx ./cmd/ceryx

# Export with: docker buildx build --platform darwin/arm64 --output bin .
FROM scratch AS binary
COPY --from=build /out/ceryx /
