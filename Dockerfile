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
    && zip /build/index.zip docker-compose.yml package.json resolv.conf


# WORKDIR /go/src/wazigate-system/

# Let's keep it in a separate layer
RUN go build -o /build/wazigate-system -i .

# Debugging stuff
# && go get github.com/go-delve/delve/cmd/dlv \  # Not supported for RPI
# COPY ./dlv.sh .
# RUN chmod +x dlv.sh 
# ENTRYPOINT [ "dlv.sh"]

ENTRYPOINT ["tail", "-f", "/dev/null"]

#----------------------------#

FROM golang:alpine AS test

WORKDIR /go/src/github.com/Waziup/wazigate-system/

ENV EXEC_PATH=/go/src/github.com/Waziup/wazigate-system/

ENTRYPOINT ["go", "test", "-v", "./..."]

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