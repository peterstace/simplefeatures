.PHONY: lint
lint:
	docker-compose -p sf-lint -f ./.ci/docker-compose-lint.yml up --abort-on-container-exit

.PHONY: unit
unit:
	docker-compose -p sf-unit -f ./.ci/docker-compose-unit.yml up --abort-on-container-exit

.PHONY: pgscan
pgscan:
	docker-compose -p sf-pgscan -f ./.ci/docker-compose-pgscan.yml up --abort-on-container-exit

.PHONY: cmppg
cmppg:
	docker-compose -p sf-cmppg -f ./.ci/docker-compose-cmppg.yml up --abort-on-container-exit

.PHONY: cmpgeos
cmpgeos:
	docker-compose -p sf-cmpgeos -f ./.ci/docker-compose-cmpgeos.yml up --abort-on-container-exit

.PHONY: geos-3.11.2
geos-3.11.2:
	env GEOS_VERSION=3.11.2 docker-compose -p sf-geos-3-11-2 -f ./.ci/docker-compose-geos.yml up --abort-on-container-exit

.PHONY: geos-3.10.5
geos-3.10.5:
	env GEOS_VERSION=3.10.5 docker-compose -p sf-geos-3-10-5 -f ./.ci/docker-compose-geos.yml up --abort-on-container-exit

.PHONY: geos-3.9.4
geos-3.9.4:
	env GEOS_VERSION=3.9.4 docker-compose -p sf-geos-3-9-4 -f ./.ci/docker-compose-geos.yml up --abort-on-container-exit

.PHONY: geos-3.8.3
geos-3.8.3:
	env GEOS_VERSION=3.8.3 docker-compose -p sf-geos-3-8-3 -f ./.ci/docker-compose-geos.yml up --abort-on-container-exit

.PHONY: geos-3.7.5
geos-3.7.5:
	env GEOS_VERSION=3.7.5 docker-compose -p sf-geos-3-7-5 -f ./.ci/docker-compose-geos.yml up --abort-on-container-exit

.PHONY: geos
geos: geos-3.11.2
geos: geos-3.10.5
geos: geos-3.9.4
geos: geos-3.8.3
geos: geos-3.7.5
