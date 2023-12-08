#!/bin/sh

BIN_DIR="/usr/local/bin/"
RES_DIR="/usr/local/etc/swmon/"

if [ ! -d $RES_DIR ];
then
      sudo mkdir $RES_DIR
fi

sudo rsync -a . $RES_DIR --exclude install.sh
sudo rsync -a swmon_mapper $BIN_DIR

(cd "$RES_DIR/.." || exit 1
  sudo chown -R root:root swmon
  sudo chmod -R 644 swmon
  sudo chmod 755 swmon
  sudo chmod 755 swmon/mibs
  sudo chmod 755 swmon/mibs/oids
  sudo chmod 755 swmon/maps
)

(cd $BIN_DIR || exit 1
  sudo chmod 755 swmon_mapper swmon_check_stp stp_mib_scanner
  sudo chown root:root swmon_mapper swmon_check_stp stp_mib_scanner
)

echo "swmon installed to $BIN_DIR"
echo "resorces, output monitoring system configuration etc. located in $RES_DIR"
echo "check README for intial configuration and further reading"