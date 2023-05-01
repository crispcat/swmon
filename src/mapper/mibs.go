package main

import utils "swmon_shared"

func LoadMibs() {

	err := utils.MIB.LoadModules()
	if err != nil {
		ErrorAll("Unable to load MIBs: %s", err)
	}

	TestLldpMibs()
}

func TestLldpMibs() {

	WriteAll("Testing LLDP MIBs persistence...")

	lldpLocPortId := MibsGetOid("LLDP-MIB::lldpLocPortId")
	WriteAll("LLDP-MIB::lldpLocPortId: %s", lldpLocPortId)

	lldpRemPortIdSubtype := MibsGetOid("LLDP-MIB::lldpRemPortIdSubtype")
	WriteAll("LLDP-MIB::lldpRemPortIdSubtype: %s", lldpRemPortIdSubtype)

	lldpRemPortId := MibsGetOid("LLDP-MIB::lldpRemPortId")
	WriteAll("LLDP-MIB::lldpRemPortId: %s", lldpRemPortId)

	lldpRemPortDesc := MibsGetOid("LLDP-MIB::lldpRemPortDesc")
	WriteAll("LLDP-MIB::lldpRemPortDesc: %s", lldpRemPortDesc)

	lldpRemSysName := MibsGetOid("LLDP-MIB::lldpRemSysName")
	WriteAll("LLDP-MIB::lldpRemSysName: %s", lldpRemSysName)

	lldpRemSysDesc := MibsGetOid("LLDP-MIB::lldpRemSysDesc")
	WriteAll("LLDP-MIB::lldpRemSysDesc: %s", lldpRemSysDesc)

	lldpRemChassisIdSubtype := MibsGetOid("LLDP-MIB::lldpRemChassisIdSubtype")
	WriteAll("LLDP-MIB::lldpRemChassisIdSubtype: %s", lldpRemChassisIdSubtype)

	lldpRemChassisId := MibsGetOid("LLDP-MIB::lldpRemChassisId")
	WriteAll("LLDP-MIB::lldpRemChassisId: %s", lldpRemChassisId)

	lldpRemSysCapSupported := MibsGetOid("LLDP-MIB::lldpRemSysCapSupported")
	WriteAll("LLDP-MIB::lldpRemSysCapSupporte: %s", lldpRemSysCapSupported)

	lldpRemSysCapEnabled := MibsGetOid("LLDP-MIB::lldpRemSysCapEnabled")
	WriteAll("LLDP-MIB::lldpRemSysCapEnabled: %s", lldpRemSysCapEnabled)
}

func MibsGetOid(name string) string {

	oid, err := utils.MIB.OID(name)
	if err != nil {
		ErrorAll("FATAL! Unable to get OID for %s", name)
	}

	return oid.String()
}
