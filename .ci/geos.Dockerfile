FROM golang:1.17-bullseye
RUN echo 'deb http://deb.debian.org/debian sid main' >> /etc/apt/sources.list && \
	apt-get -y update && \
	apt-get install -y libgeos-dev=3.11.1-1 && \
	rm -rf /var/lib/apt/lists/*
ENV PATH=/usr/lib/go-1.17/bin:${PATH}
