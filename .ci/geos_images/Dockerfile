FROM debian:bullseye AS builder

RUN apt-get update \
  && apt-get install -y cmake build-essential wget \
  && rm -rf /var/lib/apt/lists/*

ARG GEOS_VERSION=3.11.2
WORKDIR /tmp
RUN wget https://download.osgeo.org/geos/geos-$GEOS_VERSION.tar.bz2 && \
	tar -xf geos-$GEOS_VERSION.tar.bz2 && \
	rm geos-$GEOS_VERSION.tar.bz2

WORKDIR /tmp/geos-$GEOS_VERSION
RUN cmake . -DCMAKE_INSTALL_PREFIX=/usr/local && \
    make -j4 && \
    make install

FROM debian:bullseye
COPY --from=builder /usr/local/ /usr/local/
CMD ["geos-config", "--version"]
