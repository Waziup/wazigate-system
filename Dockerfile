FROM golang:alpine AS development

COPY . /go/src/github.com/Waziup/wazigate-system/
WORKDIR /go/src/github.com/Waziup/wazigate-system/
ENV GOPATH=/go/

RUN apk add --no-cache \
    git \
    iw \
    gawk \
    curl \
    gcc \
    musl-dev \
    zip \
    && mkdir /build/ \
    && cp scan.awk /build \
    && cp -r docs /build \
    && cp -r ui /build \
    && go build -o /build/wazigate-system -i . \
    && zip /build/index.zip docker-compose.yml package.json

#----------------------------#

FROM alpine:latest AS production

WORKDIR /app/
COPY --from=development /build .
RUN apk --no-cache add \
    ca-certificates \
    tzdata \
    iw \
    gawk \
    curl \
    && mv ./index.zip /

ENTRYPOINT ["./wazigate-system"]