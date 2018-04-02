FROM golang:1.10.1 AS build
COPY . /go/src/github.com/jpweber/labeler

WORKDIR /go/src/github.com/jpweber/labeler
RUN CGO_ENABLED=0 go build -a -ldflags '-s' -installsuffix  .

# copy the binary from the build stage to the final stage
FROM alpine:3.7
COPY --from=build /go/src/github.com/jpweber/labeler/labeler /usr/bin/labeler
RUN apk --update upgrade && \
    apk add ca-certificates && \
    update-ca-certificates && \
    rm -rf /var/cache/apk/*
CMD ["labeler"]