#!/bin/bash

set -eo pipefail

here="$(dirname "$(readlink -f "$0")")"

docker-compose -f "$here/docker-compose-lint.yml"    up --abort-on-container-exit
docker-compose -f "$here/docker-compose-unit.yml"    up --abort-on-container-exit
docker-compose -f "$here/docker-compose-geos.yml"    up --abort-on-container-exit
docker-compose -f "$here/docker-compose-cmpgeos.yml" up --abort-on-container-exit
docker-compose -f "$here/docker-compose-cmppg.yml"   up --abort-on-container-exit
