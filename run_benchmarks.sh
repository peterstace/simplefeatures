#!/bin/bash

set -eo pipefail

if [ $# != 2 ]; then
	echo "usage: $0 old_git_sh1 new_git_sha1"
	exit 1
fi

old_git_sha1="$1"
new_git_sha1="$2"

old="$(mktemp)"
new="$(mktemp)"
trap "rm -f $new $old" EXIT

for (( i = 0; i < 15; i++ )); do
	echo
	echo "RUN $i"

	echo
	echo "OLD"
	git checkout "$old_git_sha1"
	echo
	go test ./... -test.run='^$' -benchtime=0.1s -benchmem -bench=. | tee -a "$old"

	echo
	echo "NEW"
	git checkout "$new_git_sha1"
	echo
	go test ./... -test.run='^$' -benchtime=0.1s -benchmem -bench=. | tee -a "$new"
done

echo
echo "OLD RESULTS"
cat "$old"

echo
echo "NEW RESULTS"
cat "$new"

echo
echo "COMPARISON"

benchstat "$old" "$new"
