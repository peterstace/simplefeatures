FROM golang:1.14
RUN apt-get -y update && \
	apt-get install -y libgeos-dev=3.7.1-1 && \
	rm -rf /var/lib/apt/lists/*
ENV PATH=/usr/lib/go-1.15/bin:${PATH}
