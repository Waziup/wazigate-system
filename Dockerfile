FROM golang:1.25-bookworm AS bin

ENV GO111MODULE=on

WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=0 go build -ldflags "-s -w" -o wazigate-system .

FROM debian:bookworm-slim

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
