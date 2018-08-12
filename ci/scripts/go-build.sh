#!/bin/bash
set -e

: ${SOURCE?Need a value}

BASE=$(pwd -P)
SOURCE="$BASE"/"$SOURCE"

cd "$SOURCE"

echo "Building"
make build
