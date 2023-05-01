#!/usr/bin/env bash

ROOT_DIR=$(realpath "${0%/*}/..")

rm -rf ./release
mkdir ./release

platforms=("linux/amd64" "linux/386" "freebsd/amd64" "freebsd/386")

. "$ROOT_DIR/make/copy_resources.sh"
. "$ROOT_DIR/make/build_modules.sh"

for PLATFORM in "${platforms[@]}"
do
    PLATFORM_SPLIT=(${PLATFORM//\// })
    GOOS=${PLATFORM_SPLIT[0]}
    GOARCH=${PLATFORM_SPLIT[1]}

    if [ ! -d ./release/"${PLATFORM_SPLIT[0]}"/ ];
    then
          mkdir ./release/"${PLATFORM_SPLIT[0]}"/
          mkdir ./release/"${PLATFORM_SPLIT[0]}"/"${PLATFORM_SPLIT[1]}"
    fi

    if [ ! -d ./release/"${PLATFORM_SPLIT[0]}"/"${PLATFORM_SPLIT[1]}" ];
    then
          mkdir ./release/"${PLATFORM_SPLIT[0]}"/"${PLATFORM_SPLIT[1]}"
    fi

    env GOOS="$GOOS" GOARCH="$GOARCH"
    build_modules  "$ROOT_DIR" "$PLATFORM"
    copy_resources "$ROOT_DIR" "$PLATFORM"
done