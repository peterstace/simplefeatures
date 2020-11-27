FROM debian:bullseye
RUN apt-get -y update && \
	apt-get install -y 'libgeos-dev=3.8.1-1' 'golang-1.15' 'ca-certificates' && \
	rm -rf /var/lib/apt-lists/*
ENV PATH=/usr/lib/go-1.15/bin:${PATH}
