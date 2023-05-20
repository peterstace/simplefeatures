#!/bin/bash

set -eo pipefail

start="$(date +%s.%N)"
here="$(dirname "$(readlink -f "$0")")"
cd "$here"

docker-compose -p sf-lint -f docker-compose-lint.yml up --abort-on-container-exit
docker-compose -p sf-unit -f docker-compose-unit.yml up --abort-on-container-exit

for v in 3.11.2 3.10.5 3.9.4 3.8.3 3.7.5; do
	env GEOS_VERSION=$v docker-compose -p sf-geos-$(echo $v | tr . -) -f docker-compose-geos.yml up --abort-on-container-exit
done

docker-compose -p sf-pgscan  -f docker-compose-pgscan.yml  up --abort-on-container-exit
docker-compose -p sf-cmppg   -f docker-compose-cmppg.yml   up --abort-on-container-exit
docker-compose -p sf-cmpgeos -f docker-compose-cmpgeos.yml up --abort-on-container-exit

printf "\nduration: %.1f seconds\n" "$(echo "$(date +%s.%N) - $start" | bc)"
