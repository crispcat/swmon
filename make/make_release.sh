#!/usr/bin/env bash

ROOT_DIR=$(realpath "${0%/*}/..")

rm -rf "$ROOT_DIR"/release
mkdir "$ROOT_DIR"/release

platforms=("linux/amd64" "linux/386" "freebsd/amd64" "freebsd/386")

. "$ROOT_DIR/make/copy_resources.sh"
. "$ROOT_DIR/make/build_modules.sh"

for PLATFORM in "${platforms[@]}"
do
    PLATFORM_SPLIT=(${PLATFORM//\// })
    GOOS=${PLATFORM_SPLIT[0]}
    GOARCH=${PLATFORM_SPLIT[1]}

    RELEASE_DIR="$ROOT_DIR"/release

    if [ ! -d "$RELEASE_DIR"/"${PLATFORM_SPLIT[0]}"/ ];
    then
          mkdir "$RELEASE_DIR"/"${PLATFORM_SPLIT[0]}"/
          mkdir "$RELEASE_DIR"/"${PLATFORM_SPLIT[0]}"/"${PLATFORM_SPLIT[1]}"
    fi

    if [ ! -d "$RELEASE_DIR"/"${PLATFORM_SPLIT[0]}"/"${PLATFORM_SPLIT[1]}" ];
    then
          mkdir "$RELEASE_DIR"/"${PLATFORM_SPLIT[0]}"/"${PLATFORM_SPLIT[1]}"
    fi

    env GOOS="$GOOS" GOARCH="$GOARCH"
    build_modules  "$ROOT_DIR" "$RELEASE_DIR"/"$PLATFORM"
    copy_resources "$ROOT_DIR" "$RELEASE_DIR"/"$PLATFORM"
done