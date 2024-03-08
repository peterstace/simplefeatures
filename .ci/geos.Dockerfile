ARG ALPINE_VERSION
FROM alpine:${ALPINE_VERSION}

ARG GEOS_VERSION
RUN apk add pkgconfig gcc musl-dev geos-dev=${GEOS_VERSION}

# Alpine 3.13 doesn't include the geos.pc file, so we need to add it manually.
# It doesn't hurt to add it manually for other versions of Alpine either, so we
# just add it unconditionally.
COPY alpine_geos.pc /usr/lib/pkgconfig/geos.pc

COPY --from=golang:1.21-alpine /usr/local/go /usr/local/go
ENV PATH=${PATH}:/usr/local/go/bin
RUN go version
