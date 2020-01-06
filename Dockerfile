FROM alpine:latest AS development

COPY . /go/src/wazigate-system/
ENV GOPATH=/go/

RUN apk add --no-cache \
    go \
    git \
    iw \
    gawk \
    curl \
    gcc \
    musl-dev \
    && cd $GOPATH   \
    && go get -v -u \
        "github.com/julienschmidt/httprouter" \
        "periph.io/x/periph/conn/gpio" \
	    "periph.io/x/periph/host" \
	    "periph.io/x/periph/conn/gpio/gpioreg" \
        "periph.io/x/periph/conn/i2c/i2creg" \
        "periph.io/x/periph/devices/ssd1306" \
        "periph.io/x/periph/devices/ssd1306/image1bit" \
        "golang.org/x/image/font" \
        "golang.org/x/image/font/basicfont" \
        "golang.org/x/image/math/fixed" \  
    && mkdir /build/ \
    && cp /go/src/wazigate-system/scan.awk /build \
    && cp -r /go/src/wazigate-system/docs /build \
    && go build -o /build/wazigate-system -i /go/src/wazigate-system/

WORKDIR /go/src/wazigate-system/
ENTRYPOINT ["tail", "-f", "/dev/null"]

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