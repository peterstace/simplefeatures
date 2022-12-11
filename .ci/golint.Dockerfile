FROM golangci/golangci-lint:v1.50.1

RUN apt-get update \
  && apt-get install -y -q --no-install-recommends \
    libgeos-dev \
  && rm -rf /var/lib/apt/lists/*
