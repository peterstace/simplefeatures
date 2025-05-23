ARG ALPINE_VERSION
FROM alpine:${ALPINE_VERSION}

ARG PROJ_VERSION
RUN apk add pkgconfig gcc musl-dev proj-dev=${PROJ_VERSION}

COPY --from=golang:1.21-alpine /usr/local/go /usr/local/go
ENV PATH=${PATH}:/usr/local/go/bin
RUN go version
