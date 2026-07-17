# syntax=docker/dockerfile:1

ARG GO_VERSION=1.26.5
ARG SPEEDTEST_VERSION=1.2.0

# Download the vendor-supplied Speedtest CLI for the image target architecture.
FROM --platform=$BUILDPLATFORM debian:bookworm-slim AS speedtest
ARG TARGETARCH
ARG TARGETVARIANT
ARG SPEEDTEST_VERSION
RUN apt-get update \
    && apt-get install --no-install-recommends --yes ca-certificates curl tar \
    && case "${TARGETARCH}${TARGETVARIANT}" in \
      amd64) speedtest_arch=x86_64 ;; \
      arm64) speedtest_arch=aarch64 ;; \
      armv7) speedtest_arch=armhf ;; \
      *) echo "unsupported Speedtest architecture: ${TARGETARCH}${TARGETVARIANT}" >&2; exit 1 ;; \
    esac \
    && curl --fail --location --retry 3 --output /tmp/speedtest.tgz \
      "https://install.speedtest.net/app/cli/ookla-speedtest-${SPEEDTEST_VERSION}-linux-${speedtest_arch}.tgz" \
    && tar --extract --gzip --file /tmp/speedtest.tgz --directory /usr/local/bin speedtest \
    && chmod 0555 /usr/local/bin/speedtest \
    && rm -rf /var/lib/apt/lists/* /tmp/speedtest.tgz

FROM --platform=$BUILDPLATFORM golang:${GO_VERSION}-alpine AS builder
ARG TARGETOS
ARG TARGETARCH
ARG VERSION=dev
ARG REVISION=unknown
WORKDIR /workspace

COPY go.mod go.sum ./
RUN go mod download
COPY cmd/ cmd/
COPY internal/ internal/
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} \
    go build -trimpath \
      -ldflags="-s -w -X main.version=${VERSION} -X main.commit=${REVISION}" \
      -o /speedtest-exporter ./cmd/speedtest-exporter

# base-debian includes the glibc runtime required by Ookla's CLI; the exporter
# itself is a static Go binary. The image runs as distroless' non-root user.
FROM gcr.io/distroless/base-debian12:nonroot
COPY --from=speedtest /usr/local/bin/speedtest /usr/local/bin/speedtest
COPY --from=builder /speedtest-exporter /speedtest-exporter
USER nonroot:nonroot
EXPOSE 9798
ENTRYPOINT ["/speedtest-exporter"]
