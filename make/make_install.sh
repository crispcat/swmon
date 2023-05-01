#!/usr/bin/env bash

ROOT_DIR=$(realpath "${0%/*}/..")

. "$ROOT_DIR/make/copy_resources.sh"
. "$ROOT_DIR/make/build_modules.sh"

BUILD_DIR="$ROOT_DIR/build/"
if [ ! -d "$BUILD_DIR" ];
then
      mkdir "$BUILD_DIR"
fi

copy_resources "$ROOT_DIR" "$BUILD_DIR"
build_modules  "$ROOT_DIR" "$BUILD_DIR"

(cd "$BUILD_DIR" || exit
sudo sh install.sh)