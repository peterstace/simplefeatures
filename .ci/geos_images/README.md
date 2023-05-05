# GEOS Images

This directory contains a Dockerfile and instructions for how to create GEOS
images. These images are used for CI.

The images install GEOS from source. This is so that each version of GEOS can
be used in CI.

## Building and uploading

The following command assume the current working directory is the same as this
README.

```sh
docker build .
```
