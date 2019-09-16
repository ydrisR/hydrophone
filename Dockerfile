# Development
FROM golang:1.12-alpine AS development

RUN apk --no-cache update && \
    apk --no-cache upgrade && \
    apk add build-base git cyrus-sasl-dev rsync

ENV GO111MODULE on

WORKDIR /go/src/github.com/mdblp/hydrophone
COPY . .
RUN go get
RUN  ./build.sh

CMD ["./dist/hydrophone"]

# Release
FROM alpine:latest AS release

RUN apk --no-cache update && \
    apk --no-cache upgrade && \
    apk add --no-cache ca-certificates && \
	apk add --no-cache libsasl	&& \
    adduser -D mdblp

WORKDIR /home/mdblp

USER mdblp

COPY --from=development --chown=mdblp /go/src/github.com/mdblp/hydrophone/dist/hydrophone .
COPY --chown=mdblp templates/html ./templates/html/
COPY --chown=mdblp templates/locales ./templates/locales/
COPY --chown=mdblp templates/meta ./templates/meta/

CMD ["./hydrophone"]
