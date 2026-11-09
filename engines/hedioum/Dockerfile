# syntax=docker/dockerfile:1
#
# sober-tunnel camouflage engine — minimal, static, multi-arch image.
#
# Build (single arch, host arch):
#   docker build -t sober-tunnel-mimic:local .
#
# Build (multi-arch, requires buildx):
#   docker buildx build --platform linux/amd64,linux/arm64,linux/arm/v7 \
#       -t sober-tunnel-mimic:local --push .
#
# Run the Iran hub with TUN mode (needs NET_ADMIN + the tun device):
#   docker run -d --name sober-tunnel-mimic --restart unless-stopped \
#       --cap-add NET_ADMIN --device /dev/net/tun \
#       -v /etc/sober-tunnel/standalone:/etc/sober-tunnel/standalone \
#       -p 127.0.0.1:40001:40001 sober-tunnel-mimic:local
#
# SOCKS-only (no TUN) needs neither the cap nor the device.
# One-off config, writing into the mounted volume:
#   docker run --rm -v /etc/sober-tunnel/standalone:/etc/sober-tunnel/standalone sober-tunnel-mimic:local \
#       setup-iran --alias FR --token <pairing-token> --socks-port 40001 --tun --dns

# ---- build stage ----
FROM --platform=$BUILDPLATFORM golang:1.26.10 AS build
ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT
WORKDIR /src

# Cache modules first.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# GOARM is only consulted for arm/v7; harmless otherwise.
RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    GOARM=$(echo "${TARGETVARIANT}" | tr -d 'v') \
    go build -trimpath -ldflags="-s -w -X main.Date=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
        -o /out/sober-tunnel-mimic ./cmd/hedioum

# ---- runtime stage ----
FROM scratch
# CA roots so the TLS mimic / egress can validate real certificates.
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /out/sober-tunnel-mimic /usr/local/bin/sober-tunnel-mimic
VOLUME ["/etc/sober-tunnel/standalone"]
ENTRYPOINT ["/usr/local/bin/sober-tunnel-mimic"]
