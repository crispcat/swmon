#!/usr/bin/env bash

BIN_DIR="/usr/local/bin/"
RES_DIR="/usr/local/etc/swmon/"

echo "removing resources at $RES_DIR"
sudo rm -rf /usr/local/etc/swmon/

echo "removing binaries in $BIN_DIR"
sudo rm -f /usr/local/bin/swmon_mapper
sudo rm -f /usr/local/bin/swmon_check_stp
sudo rm -f /usr/local/bin/stp_mib_scanner

echo "done"