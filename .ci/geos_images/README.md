# GEOS Images

This directory contains a Dockerfile and instructions for how to create GEOS
images. These images are used for CI.

The images install GEOS from source. This is so that each version of GEOS can
be used in CI.

## Building and uploading

Assumes using an `amd64` AWS EC2 instance with Ubuntu Linux. The instance size
should be compute optimised, e.g. `c6a.xlarge`.

Install `docker`:
```sh
sudo snap install docker
sudo groupadd docker
sudo usermod -aG docker $USER
sudo reboot
docker run hello-world
```

Set up buildx:
```sh
docker buildx create --name mybuilder
docker buildx use mybuilder
```

Login to dockerhub:
```sh
docker login # interactive
```

Specify the versions of GEOS and Go to build the images for:
```sh
GEOS_VERSION=3.10.5
GO_VERSION=1.20.4
```

Build and push the GEOS images:
```sh
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  --file geos.Dockerfile \
  --build-arg GEOS_VERSION=${GEOS_VERSION} \
  --tag peterstace/simplefeatures-ci:geos-${GEOS_VERSION} \
  --push .
```

Build and push the Go images:
```sh
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  --file go.Dockerfile \
  --build-arg GEOS_VERSION=${GEOS_VERSION} \
  --build-arg GO_VERSION=${GO_VERSION} \
  --tag peterstace/simplefeatures-ci:geos-${GEOS_VERSION}-go-${GO_VERSION} \
  --push .
```
