package main

import (
	"github.com/hallidave/mibtool/smi"
)

const ETC = "/usr/local/etc/swmon/"
const MIBS_ROOT = ETC + "mibs"

var MIB = smi.NewMIB(MIBS_ROOT)
