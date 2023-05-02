#!/usr/bin/env bash

VERSION="0.1"

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
    mkdir "$RELEASE_DIR"/"${PLATFORM_SPLIT[0]}"/
    mkdir "$RELEASE_DIR"/"${PLATFORM_SPLIT[0]}"/"${PLATFORM_SPLIT[1]}"
    mkdir "$RELEASE_DIR"/"${PLATFORM_SPLIT[0]}"/"${PLATFORM_SPLIT[1]}"/swmon

    env GOOS="$GOOS" GOARCH="$GOARCH"
    build_modules  "$ROOT_DIR" "$RELEASE_DIR"/"$PLATFORM"/swmon/
    copy_resources "$ROOT_DIR" "$RELEASE_DIR"/"$PLATFORM"/swmon/

    (cd "$RELEASE_DIR"/"$PLATFORM" || exit 1
    tar -czvf "swmon_${PLATFORM_SPLIT[0]}_${PLATFORM_SPLIT[1]}_$VERSION.tar.gz" ./swmon
    mv "swmon_${PLATFORM_SPLIT[0]}_${PLATFORM_SPLIT[1]}_$VERSION.tar.gz"  "$ROOT_DIR"/release/)
done