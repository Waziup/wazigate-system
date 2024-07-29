################################################################################


FROM golang:1.16-alpine AS bin

ENV CGO_ENABLED=0
ENV GO111MODULE=on

RUN apk add --no-cache ca-certificates tzdata git

WORKDIR /app/

COPY go.mod go.sum /app/
RUN go mod download

COPY . /app
RUN go build -ldflags "-s -w" -o wazigate-system .

################################################################################


FROM alpine:latest AS app

RUN apk add --no-cache iw gawk ca-certificates tzdata curl

WORKDIR /app/

COPY docs /app/docs

COPY --from=bin /app/wazigate-system .

ENV WAZIUP_MONGO=wazigate-mongo:27017

HEALTHCHECK CMD curl --fail http://localhost || exit 1 

VOLUME /var/lib/waziapp

ENTRYPOINT ["./wazigate-system"]
