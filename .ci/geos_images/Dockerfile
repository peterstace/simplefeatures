FROM debian:bullseye

RUN apt-get update \
	&& apt-get install -y cmake build-essential wget \
	&& rm -rf /var/lib/apt/lists/*

ARG GEOS_VERSION
RUN cd /tmp \
	&& wget https://download.osgeo.org/geos/geos-$GEOS_VERSION.tar.bz2 \
	&& tar -xf geos-$GEOS_VERSION.tar.bz2 \
	&& rm geos-$GEOS_VERSION.tar.bz2 \
	&& cd geos-$GEOS_VERSION \
	&& cmake . -DCMAKE_INSTALL_PREFIX=/usr/local \
	&& make -j$(nproc) \
	&& make install \
	&& ldconfig \
	&& cd .. \
	&& rm -r geos-$GEOS_VERSION

ARG GO_VERSION
ARG TARGETARCH
RUN cd /tmp \
	&& wget https://go.dev/dl/go$GO_VERSION.linux-$TARGETARCH.tar.gz \
	&& tar -C /usr/local -xzf go$GO_VERSION.linux-$TARGETARCH.tar.gz \
	&& rm go$GO_VERSION.linux-$TARGETARCH.tar.gz
ENV PATH=$PATH:/usr/local/go/bin

CMD geos-config --version && go version
