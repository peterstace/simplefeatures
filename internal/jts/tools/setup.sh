#!/usr/bin/env bash
#
# Populates tools/.cache with JTS sources at the ported release, the JUnit jars
# needed to type-check the JTS tests, and compiled JTS core classes for the
# oracle.

set -euo pipefail

JTS_VERSION=1.20.0
JUNIT_VERSION=4.13.1
JUNIT_SHA1=cdd00374f1fee76b11e2a9d127405aa3f6be5b6a
HAMCREST_VERSION=1.3
HAMCREST_SHA1=42a25dc3219429f0e5d060061f71acb49bf010a0

tools_dir="$(cd "$(dirname "$0")" && pwd)"
cache="$tools_dir/.cache"
mkdir -p "$cache"

fetch_jar() {
	local url="$1" sha1="$2" dest="$3"
	if [[ ! -f "$dest" ]]; then
		curl -fsSL -o "$dest.tmp" "$url"
		echo "$sha1  $dest.tmp" | sha1sum -c --quiet
		mv "$dest.tmp" "$dest"
	fi
}

src="$cache/jts-$JTS_VERSION"
if [[ ! -d "$src" ]]; then
	tmp="$(mktemp -d)"
	curl -fsSL "https://github.com/locationtech/jts/archive/refs/tags/$JTS_VERSION.tar.gz" |
		tar -xz -C "$tmp"
	mv "$tmp/jts-$JTS_VERSION" "$src"
	rmdir "$tmp"
fi

fetch_jar "https://repo1.maven.org/maven2/junit/junit/$JUNIT_VERSION/junit-$JUNIT_VERSION.jar" \
	"$JUNIT_SHA1" "$cache/junit.jar"
fetch_jar "https://repo1.maven.org/maven2/org/hamcrest/hamcrest-core/$HAMCREST_VERSION/hamcrest-core-$HAMCREST_VERSION.jar" \
	"$HAMCREST_SHA1" "$cache/hamcrest-core.jar"

classes="$cache/jts-classes"
if [[ ! -d "$classes" ]]; then
	find "$src/modules/core/src/main/java" -name '*.java' >"$cache/jts-sources.txt"
	javac --release 8 -nowarn -encoding UTF-8 -d "$classes.tmp" @"$cache/jts-sources.txt" 2>&1 |
		grep -v -e '^Note:' -e 'bootstrap classpath' -e '^warning: \[options\]' -e '^[0-9]* warning' || true
	mv "$classes.tmp" "$classes"
fi
