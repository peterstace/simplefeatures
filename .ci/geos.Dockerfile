FROM golang:1.19-buster

RUN apt-get -y update && \
	apt-get install -y bzip2 cmake && \
	rm -rf /var/lib/apt/lists/*
ENV PATH=/usr/lib/go-1.17/bin:${PATH}

ARG GEOS_VERSION=3.11.2

# Instructions from https://libgeos.org/usage/download/
RUN wget https://download.osgeo.org/geos/geos-${GEOS_VERSION}.tar.bz2 -o geos.tar.bz2
RUN tar xvfj geos-${GEOS_VERSION}.tar.bz2
RUN mkdir -p geos-${GEOS_VERSION}/_build
WORKDIR geos-${GEOS_VERSION}/_build
RUN cmake -DCMAKE_BUILD_TYPE=Release -DCMAKE_INSTALL_PREFIX=/usr/local ..
RUN make
RUN ctest
RUN make install

ENV LD_LIBRARY_PATH=/usr/local/lib

#RUN apt-get -y update
#RUN apt-get install -y software-properties-common
#RUN add-apt-repository ppa:ubuntugis/ppa
#RUN apt-get -y update
#RUN apt-get install geos
#RUN apt-get -y update && \
#	apt-get install -y libgeos-dev=3.11.1-1 && \
#	rm -rf /var/lib/apt/lists/*
#ENV PATH=/usr/lib/go-1.17/bin:${PATH}
