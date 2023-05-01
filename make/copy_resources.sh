#!/usr/bin/env bash

copy_resources() {
    local SOURCE_DIR=$1
    local BUILD_DIR=$2
    # copy mibs
    rsync -a "$SOURCE_DIR/mibs" "$BUILD_DIR/"
    # copy templates
    rsync -a "$SOURCE_DIR/maps" "$BUILD_DIR/"
    # copy installator
    rsync -a "$SOURCE_DIR/make/install.sh" "$BUILD_DIR/"
}