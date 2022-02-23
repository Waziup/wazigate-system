FROM waziup/node-sass:14 AS ui

ARG MODE=prod

WORKDIR /app/

COPY ui/package.json /app/
RUN npm i

COPY ui/. /app
RUN npm run build --env $MODE

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

COPY --from=ui /app/node_modules/react/umd ui/node_modules/react/umd
COPY --from=ui /app/node_modules/react-dom/umd ui/node_modules/react-dom/umd
COPY --from=ui /app/index.html /app/dev.html /app/favicon.ico ui/
COPY --from=ui /app/dist ui/dist
COPY --from=ui /app/icons ui/icons

COPY docs /app/docs

COPY --from=bin /app/wazigate-system .

COPY scan.awk .

ENV WAZIUP_MONGO=wazigate-mongo:27017

HEALTHCHECK CMD curl --fail http://localhost || exit 1 

VOLUME /var/lib/waziapp

ENTRYPOINT ["./wazigate-system"]
