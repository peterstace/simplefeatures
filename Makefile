.PHONY: all
all: unit lint geos pgscan cmppg cmpgeos

DC_RUN = \
	docker compose \
	--project-name sf-$$task \
	--file .ci/compose-$$task.yaml \
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

.PHONY: scrapestrings
scrapestrings:
	go run ./internal/cmprefimpl/scraper

DC_GEOS_RUN = \
	docker compose \
	--project-name sf-geos-$$(echo $$geos_version | sed 's/\./-/g') \
	--file .ci/compose-geos.yaml \
	up \
	--build \
	--abort-on-container-exit

DC_PROJ_RUN = \
	docker compose \
	--project-name sf-proj-$$(echo $$proj_version | sed 's/\./-/g') \
	--file .ci/compose-proj.yaml \
	up \
	--build \
	--abort-on-container-exit

.PHONY: geos-3.12
geos-3.12:
	export tags='' alpine_version=3.19 geos_version=3.12.1-r0; $(DC_GEOS_RUN)

.PHONY: geos-3.11
geos-3.11:
	export tags='' alpine_version=3.18 geos_version=3.11.2-r0; $(DC_GEOS_RUN)

.PHONY: geos-3.10
geos-3.10:
	export tags='' alpine_version=3.16 geos_version=3.10.3-r0; $(DC_GEOS_RUN)

.PHONY: geos-3.9
geos-3.9:
	export tags='' alpine_version=3.14 geos_version=3.9.1-r0; $(DC_GEOS_RUN)

.PHONY: geos-3.8
geos-3.8:
	# Alpine 3.13 doesn't include a geos.pc file (needed by pkg-config). So
	# the sfnopkgconfig tag is used, disabling the use of pkg-config.
	# LDFLAGS are used to configure GEOS directly.
	export tags='-tags sfnopkgconfig' \
	       alpine_version=3.13 geos_version=3.8.1-r2; $(DC_GEOS_RUN)

.PHONY: geos
geos: geos-3.12 geos-3.11 geos-3.10 geos-3.9 geos-3.8

.PHONY: proj-9.5
proj-9.5:
	export tags='' alpine_version=3.21 proj_version=9.5.0-r0; $(DC_PROJ_RUN)

.PHONY: proj-9.4
proj-9.4:
	export tags='' alpine_version=3.20 proj_version=9.4.0-r0; $(DC_PROJ_RUN)

.PHONY: proj-9.3
proj-9.3:
	export tags='' alpine_version=3.19 proj_version=9.3.1-r0; $(DC_PROJ_RUN)

.PHONY: proj-9.2
proj-9.2:
	export tags='' alpine_version=3.18 proj_version=9.2.1-r0; $(DC_PROJ_RUN)

.PHONY: proj-9.1
proj-9.1:
	export tags='' alpine_version=3.17 proj_version=9.1.0-r0; $(DC_PROJ_RUN)

.PHONY: proj-9.0
proj-9.0:
	export tags='' alpine_version=3.16 proj_version=9.0.0-r0; $(DC_PROJ_RUN)

.PHONY: proj-8.2
proj-8.2:
	export tags='' alpine_version=3.15 proj_version=8.2.0-r0; $(DC_PROJ_RUN)

.PHONY: proj-7.2
proj-7.2:
	export tags='' alpine_version=3.14 proj_version=7.2.1-r0; $(DC_PROJ_RUN)

.PHONY: proj-7.1
proj-7.1:
	export tags='' alpine_version=3.13 proj_version=7.1.1-r0; $(DC_PROJ_RUN)

.PHONY: proj
proj: proj-9.5 proj-9.4 proj-9.3 proj-9.2 proj-9.1 proj-9.0 proj-8.2 proj-7.2 proj-7.1
