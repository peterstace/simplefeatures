.PHONY: all
all: unit lint geos pgscan cmppg cmpgeos

DC_RUN = \
	docker compose \
	--project-name sf-$$task \
	--file .ci/docker-compose-$$task.yml \
	up \
	--abort-on-container-exit \
	--build

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
	docker compose \
	--project-name sf-geos-$$(echo $$geos_version | sed 's/\./-/g') \
	--file .ci/docker-compose-geos.yml \
	up \
	--build \
	--abort-on-container-exit

.PHONY: geos-3.12
geos-3.12:
	export alpine_version=3.19 geos_version=3.12.1-r0; $(DC_GEOS_RUN)

.PHONY: geos-3.11
geos-3.11:
	export alpine_version=3.18 geos_version=3.11.2-r0; $(DC_GEOS_RUN)

.PHONY: geos-3.10
geos-3.10:
	export alpine_version=3.16 geos_version=3.10.3-r0; $(DC_GEOS_RUN)

.PHONY: geos-3.9
geos-3.9:
	export alpine_version=3.14 geos_version=3.9.1-r0; $(DC_GEOS_RUN)

.PHONY: geos-3.8
geos-3.8:
	# Alpine 3.13 doesn't include a geos.pc file (needed by pkg-config). So
	# the sfnopkgconfig tag is used, disabling the use of pkg-config.
	# LDFLAGS are used to configure GEOS directly.
	export tags='-tags sfnopkgconfig' \
	       alpine_version=3.13 geos_version=3.8.1-r2; $(DC_GEOS_RUN)

.PHONY: geos
geos: geos-3.12 geos-3.11 geos-3.10 geos-3.9 geos-3.8
