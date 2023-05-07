ARG GEOS_VERSION

# TODO: choose between arm64 and amd64 dynamically.

FROM debian:bullseye AS builder
RUN apt-get update \
	&& apt-get install -y wget \
	&& rm -rf /var/lib/apt/lists/*
WORKDIR /tmp
ARG GO_VERSION
RUN mkdir -p /tmp/usr/local \
	&& wget https://go.dev/dl/go${GO_VERSION}.linux-arm64.tar.gz \
	&& tar -C /tmp/usr/local -xzf go${GO_VERSION}.linux-arm64.tar.gz

FROM peterstace/simplefeatures-ci:geos-$GEOS_VERSION
ARG GEOS_VERSION
COPY --from=builder /tmp/usr/local /usr/local
ENV PATH=$PATH:/usr/local/go/bin
CMD geos-config --version && go version
