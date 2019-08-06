FROM golang:1.12-alpine AS build-env

LABEL maintainer="Daniel Hebestreit"

RUN apk update && apk add --no-cache libc6-compat ca-certificates git && rm -rf /var/cache/apk/*

ENV CGO_ENABLED 0
RUN go get -u github.com/golang/dep/...

WORKDIR $GOPATH/src/github.com/hebestreit/spotify-headphone-party
COPY Gopkg.toml Gopkg.lock ./
RUN dep ensure --vendor-only

COPY . ./
RUN GOOS=linux go build -gcflags "all=-N -l" -a -installsuffix nocgo -o /spotify-headphone-party .


FROM alpine:3.7

LABEL maintainer="Daniel Hebestreit"

RUN apk update && apk add --no-cache ca-certificates && rm -rf /var/cache/apk/*

WORKDIR /

COPY --from=build-env /spotify-headphone-party /
COPY ./template ./template

EXPOSE 8090

CMD ["/spotify-headphone-party"]