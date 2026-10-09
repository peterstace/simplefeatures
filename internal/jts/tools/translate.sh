#!/usr/bin/env bash
#
# Regenerates the Go translation of JTS in internal/jts.

set -euo pipefail

tools_dir="$(cd "$(dirname "$0")" && pwd)"
jts_dir="$(dirname "$tools_dir")"
cache="$tools_dir/.cache"
"$tools_dir/setup.sh"

exports=()
for pkg in api code tree util file comp; do
	exports+=(--add-exports "jdk.compiler/com.sun.tools.javac.$pkg=ALL-UNNAMED")
done

classes="$cache/translator-classes"
rm -rf "$classes"
javac -nowarn -Xlint:none "${exports[@]}" -d "$classes" "$tools_dir"/translator/src/*.java
java -Xss64m "${exports[@]}" -cp "$classes" Main \
	--jts "$cache/jts-1.20.0/modules/core/src" \
	--classpath "$cache/junit.jar:$cache/hamcrest-core.jar" \
	--out "$jts_dir" \
	--reports "$tools_dir/translator"

# The full list of runtime members referenced, ignoring what is missing.
java -Xss64m "${exports[@]}" -cp "$classes" Main \
	--jts "$cache/jts-1.20.0/modules/core/src" \
	--classpath "$cache/junit.jar:$cache/hamcrest-core.jar" \
	--out "$jts_dir" \
	--api-report "$tools_dir/translator/runtime-api.txt"
