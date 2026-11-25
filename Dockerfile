# Pinned to an exact patch, and to trixie to match wazigate-lora and the runtime
# stage below. 1.25-bookworm was a floating patch tag -- the same hazard class as
# the golang:trixie that silently moved mid-build. CGO is off, so the builder
# distro does not affect the output binary. go.mod declares go 1.16 and carries
# no toolchain directive, so the language version is unchanged by this bump.
FROM golang:1.26.6-trixie AS bin

ENV GO111MODULE=on

WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .

# Go's signal-based async preemption (SIGURG) can livelock under qemu-user,
# pinning one core at 100% instead of compiling. Costs nothing natively.
ENV GODEBUG=asyncpreemptoff=1

RUN CGO_ENABLED=0 go build -ldflags "-s -w" -o wazigate-system .

# Aligned with the rest of the project and with the gateway OS: the ISO is now
# Debian trixie and wazigate-lora-forwarders already ships debian:trixie-slim.
# All of the packages below exist in trixie main. Revert to ubuntu:24.04 if
# NetworkManager behaviour differs from what wazigate-system expects.
FROM debian:trixie-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
    curl iw gawk tzdata network-manager network-manager-openvpn \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/* /usr/share/doc /usr/share/man


WORKDIR /app/

COPY docs /app/docs

COPY --from=bin /app/wazigate-system .

ENV WAZIUP_MONGO=wazigate-mongo:27017

HEALTHCHECK CMD curl --fail http://localhost || exit 1 

VOLUME /var/lib/waziapp

ENTRYPOINT ["./wazigate-system"]
