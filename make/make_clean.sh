#!/usr/bin/env bash

ROOT_DIR=$(realpath "${0%/*}/..")

rm -rf "$ROOT_DIR/build"
rm -rf "$ROOT_DIR/release"
