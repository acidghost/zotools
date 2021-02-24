#!/usr/bin/env bash

set -e

if [ -z "$VERSION" ]
then
    VERSION="v1.2.1"
fi

build_dir=$(mktemp -d)
git clone https://github.com/bats-core/bats-core "$build_dir"

dir=$(pwd)
pushd "$build_dir"
git checkout --detach "$VERSION"
./install.sh "$dir/test/bats"
popd

rm -rf "$build_dir"
