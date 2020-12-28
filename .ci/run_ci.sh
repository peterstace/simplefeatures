#!/bin/bash

set -eo pipefail

start="$(date +%s.%N)"
here="$(dirname "$(readlink -f "$0")")"

docker-compose -f "$here/.ci/docker-compose-lint.yml"    up --abort-on-container-exit
docker-compose -f "$here/.ci/docker-compose-unit.yml"    up --abort-on-container-exit
docker-compose -f "$here/.ci/docker-compose-geos.yml"    up --abort-on-container-exit
docker-compose -f "$here/.ci/docker-compose-cmpgeos.yml" up --abort-on-container-exit
docker-compose -f "$here/.ci/docker-compose-cmppg.yml"   up --abort-on-container-exit

printf "\nduration: %.1f seconds\n" "$(echo "$(date +%s.%N) - $start" | bc)"
