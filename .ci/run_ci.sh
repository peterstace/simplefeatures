#!/bin/bash

set -eo pipefail

start="$(date +%s.%N)"
here="$(dirname "$(readlink -f "$0")")"

docker-compose -p sf-lint    -f "$here/docker-compose-lint.yml"    up --abort-on-container-exit
docker-compose -p sf-unit    -f "$here/docker-compose-unit.yml"    up --abort-on-container-exit
docker-compose -p sf-geos    -f "$here/docker-compose-geos.yml"    up --abort-on-container-exit
docker-compose -p sf-pgscan  -f "$here/docker-compose-pgscan.yml"  up --abort-on-container-exit
docker-compose -p sf-cmppg   -f "$here/docker-compose-cmppg.yml"   up --abort-on-container-exit
docker-compose -p sf-cmpgeos -f "$here/docker-compose-cmpgeos.yml" up --abort-on-container-exit

printf "\nduration: %.1f seconds\n" "$(echo "$(date +%s.%N) - $start" | bc)"
