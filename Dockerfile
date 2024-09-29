# Use the official Go image as the base image
FROM golang:1.21-alpine AS builder

# Version to build. Default is the Git HEAD.
ARG VERSION="HEAD"
# Use muslc for static libs
ARG BUILD_TAGS="muslc"

RUN apk add --no-cache --update openssh git make build-base linux-headers \
    pkgconfig zeromq-dev libsodium-dev \
    libzmq-static libsodium-static gcc

RUN mkdir -p /root/.ssh && ssh-keyscan github.com >> /root/.ssh/known_hosts
RUN git config --global url."git@github.com:".insteadOf "https://github.com/"
ENV GOPRIVATE=github.com/babylonlabs-io/*

# Build
WORKDIR /go/src/github.com/babylonlabs-io/finality-gadget
# Cache dependencies
COPY go.mod go.sum /go/src/github.com/babylonlabs-io/finality-gadget/
RUN go mod download
# Copy the rest of the files
COPY ./ /go/src/github.com/babylonlabs-io/finality-gadget/

# Download the correct libwasmvm version for the static linking with the build tag 'muslc'
RUN WASMVM_VERSION=$(go list -m github.com/CosmWasm/wasmvm/v2 | cut -d ' ' -f 2) && \
    wget https://github.com/CosmWasm/wasmvm/releases/download/$WASMVM_VERSION/libwasmvm_muslc.$(uname -m).a \
        -O /lib/libwasmvm_muslc.$(uname -m).a && \
    # verify checksum
    wget https://github.com/CosmWasm/wasmvm/releases/download/$WASMVM_VERSION/checksums.txt -O /tmp/checksums.txt && \
    sha256sum /lib/libwasmvm_muslc.$(uname -m).a | grep $(cat /tmp/checksums.txt | grep libwasmvm_muslc.$(uname -m) | cut -d ' ' -f 1)

RUN CGO_LDFLAGS="$CGO_LDFLAGS -lstdc++ -lm -lsodium" \
    CGO_ENABLED=1 \
    BUILD_TAGS=$BUILD_TAGS \
    LINK_STATICALLY=true \
    make build

# FINAL IMAGE
FROM alpine:3.16

RUN addgroup --gid 1138 -S finality-gadget && adduser --uid 1138 -S finality-gadget -G finality-gadget

RUN apk add bash curl jq libstdc++ libgcc

COPY --from=builder /go/src/github.com/babylonlabs-io/finality-gadget/build/opfgd /bin/opfgd

WORKDIR /home/finality-gadget
RUN chown -R finality-gadget /home/finality-gadget
USER finality-gadget