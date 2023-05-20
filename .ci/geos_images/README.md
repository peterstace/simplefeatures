# GEOS Images

This directory contains a Dockerfile and instructions for how to create GEOS
images. These images are used for CI.

The images install GEOS from source. This is so that historic versions GEOS can
be used in CI, testing backwards compatibility with old releases.

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

Use tmux in case the SSH connection goes down (optional):
```sh
tmux
```

Specify the versions of GEOS and Go to build the images for:
```sh
GEOS_VERSION=3.11.2
GO_VERSION=1.20.4
```

Build and push the image:
```sh
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  --file Dockerfile \
  --build-arg GEOS_VERSION=${GEOS_VERSION} \
  --build-arg GO_VERSION=${GO_VERSION} \
  --tag peterstace/simplefeatures-ci:geos-${GEOS_VERSION}-go-${GO_VERSION} \
  --push .
```
