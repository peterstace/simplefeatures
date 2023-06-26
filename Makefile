DC_RUN = \
	docker compose \
	--project-name sf-$$task \
	--file .ci/docker-compose-$$task.yml \
	up \
	--abort-on-container-exit

.PHONY: lint
lint:
	task=lint; $(DC_RUN)

.PHONY: unit
unit:
	task=unit; $(DC_RUN)

.PHONY: pgscan
pgscan:
	task=pgscan; $(DC_RUN)

.PHONY: cmppg
cmppg:
	task=cmppg; $(DC_RUN)

.PHONY: cmpgeos
cmpgeos:
	task=cmpgeos; $(DC_RUN)

DC_GEOS_RUN = \
	env GEOS_VERSION=$$version \
	docker compose \
	--project-name sf-geos-$$(echo $$version | sed 's/\./-/g') \
	--file .ci/docker-compose-geos.yml \
	up \
	--abort-on-container-exit

.PHONY: geos-3.11
geos-3.11:
	version=3.11.2; $(DC_GEOS_RUN)

.PHONY: geos-3.10
geos-3.10:
	version=3.10.5; $(DC_GEOS_RUN)

.PHONY: geos-3.9
geos-3.9:
	version=3.9.4; $(DC_GEOS_RUN)

.PHONY: geos-3.8
geos-3.8:
	version=3.8.3; $(DC_GEOS_RUN)

.PHONY: geos-3.7
geos-3.7:
	version=3.7.5; $(DC_GEOS_RUN)

.PHONY: geos
geos: geos-3.11
geos: geos-3.10
geos: geos-3.9
geos: geos-3.8
geos: geos-3.7
