#!/bin/bash
set -e

: ${SOURCE?Need a value}
: ${CODE_DIR?Need a value}

BASE=$(pwd -P)
SOURCE="$BASE"/"$SOURCE"
PACKAGE_PATH="$GOPATH"/src/github.com/terraform-providers/terraform-provider-aws

mkdir -p "$PACKAGE_PATH"

cp -r "$SOURCE" "$PACKAGE_PATH"

cd "$PACKAGE_PATH"

echo "Building"
make build
