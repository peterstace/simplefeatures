ARG GEOS_VERSION

FROM debian:bullseye AS builder
RUN apt-get update \
	&& apt-get install -y wget \
	&& rm -rf /var/lib/apt/lists/*
WORKDIR /tmp
ARG GO_VERSION
ARG TARGETARCH
RUN wget https://go.dev/dl/go$GO_VERSION.linux-$TARGETARCH.tar.gz \
	&& tar -C /usr/local -xzf go$GO_VERSION.linux-$TARGETARCH.tar.gz

FROM peterstace/simplefeatures-ci:geos-$GEOS_VERSION
COPY --from=builder /usr/local /usr/local
ENV PATH=$PATH:/usr/local/go/bin
CMD geos-config --version && go version
