#!/usr/bin/env bash

build_modules () {
    local SOURCE_DIR=$1
    local BUILD_DIR=$2
    # build modules
    build_module "$SOURCE_DIR/src/mapper" "$BUILD_DIR"
    build_module "$SOURCE_DIR/src/mib-scanner" "$BUILD_DIR"
    build_module "$SOURCE_DIR/src/nagios-plugins/stp-plugin" "$BUILD_DIR"
}

build_module () {
    local SOURCE_DIR=$1
    local BUILD_DIR=$2
    echo "Building module $SOURCE_DIR to $BUILD_DIR"
    (cd "$SOURCE_DIR" || (echo "Cannot cd to $SOURCE_DIR"; exit 1)
        go mod tidy
        go build -o "$BUILD_DIR"
        if [ $? -ne 0 ]; then
               echo 'An error has occurred! Aborting the build.'
               exit 1
    fi)
}