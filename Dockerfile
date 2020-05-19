FROM alpine:latest AS development

COPY . /go/src/github.com/Waziup/wazigate-system/
WORKDIR /go/src/github.com/Waziup/wazigate-system/
ENV GOPATH=/go/

RUN apk add --no-cache \
    go \
    git \
    iw \
    gawk \
    curl \
    gcc \
    musl-dev \
    && mkdir /build/ \
    && cp scan.awk /build \
    && cp -r docs /build \
    && go build -o /build/wazigate-system -i . 

#----------------------------#

FROM alpine:latest AS production

WORKDIR /app/
COPY --from=development /build .
RUN apk --no-cache add \
    ca-certificates \
    tzdata \
    iw \
    gawk \
    curl

ENTRYPOINT ["./wazigate-system"]
