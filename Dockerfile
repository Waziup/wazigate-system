FROM golang:alpine3.12 AS development

COPY . /go/src/github.com/Waziup/wazigate-system/
WORKDIR /go/src/github.com/Waziup/wazigate-system/
ENV GOPATH=/go/

RUN sed -i 's/https/http/' /etc/apk/repositories

RUN apk add --no-cache \
    git \
    iw \
    gawk \
    curl \
    gcc \
    musl-dev \
    zip \
    && mkdir -p /build/ \
    && cp scan.awk /build \
    && cp -r docs /build \
    # && cp -r ui /build \
    && zip /build/index.zip docker-compose.yml package.json resolv.conf


# Copy the UI Files
COPY ui/node_modules/react/umd /build/ui/node_modules/react/umd
COPY ui/node_modules/react-dom/umd /build/ui/node_modules/react-dom/umd
COPY ui/index.html \
    ui/favicon.ico \
    /build/ui/
COPY ui/dist /build/ui/dist
COPY ui/icons /build/ui/icons


# WORKDIR /go/src/wazigate-system/

# Let's keep it in a separate layer
RUN go build -o /build/wazigate-system .

# Debugging stuff
# && go get github.com/go-delve/delve/cmd/dlv \  # Currently NOT supported for RPI
# COPY ./dlv.sh .
# RUN chmod +x dlv.sh 
# ENTRYPOINT [ "dlv.sh"]

ENTRYPOINT ["tail", "-f", "/dev/null"]

#----------------------------#

FROM development AS test

WORKDIR /go/src/github.com/Waziup/wazigate-system/

ENV EXEC_PATH=/go/src/github.com/Waziup/wazigate-system/

ENTRYPOINT ["go", "test", "-v", "./..."]

#----------------------------#

# FROM alpine:latest AS production
FROM alpine:3.14 AS production

WORKDIR /app/
COPY --from=development /build .

RUN sed -i 's/https/http/' /etc/apk/repositories

RUN apk --no-cache add \
    ca-certificates \
    tzdata \
    iw \
    gawk \
    curl \
    && mv ./index.zip /

ENTRYPOINT ["./wazigate-system"]