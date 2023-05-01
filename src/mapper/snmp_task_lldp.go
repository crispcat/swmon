package main

import utils "swmon_shared"

func SNMP_GetLLDP(task *NetTask, queue *NetTaskQueue, hostsMap *HostsModel) {

	client, finalizer, err := SnmpStartRoutine(task, queue)
	if err != nil {
		return
	}

	defer finalizer()

	WriteAll("[%s] Starting LLDP probes...", task.ip)

	host := hostsMap.GetOrCreate(task.ip)

	// we will supress errors to retrieve all data we can

	locPortIdReq := CraftSnmpRequest(host, "LLDP-MIB::lldpLocPortId", ResultTypeAuto)
	_ = WalkOids(client, []*SnmpRequest{locPortIdReq}, func(req *SnmpRequest) error {

		portNumber, ok := AssumePortNumberOnPos(req, req.GetSplittedOid(), -1)
		if !ok {
			return nil
		}

		locPortId := req.ParseSnmpResult()
		if locPortId == "" {
			WriteAll("[%s]LLDP-MIB::%s [%s] REQEST: LOCAL PORT ID IS EMPTY! MAYBE UNSUPPORTED OID SCHEME",
				host.Ip, req.name, req.oid)
		}

		localLldpPort := host.GetOrCreateLocalLldpPort(portNumber)
		localLldpPort.Id = utils.Sanitize(locPortId)

		return nil
	})

	remPortId := CraftSnmpRequest(host, "LLDP-MIB::lldpRemPortId", ResultTypeAuto)
	_ = WalkOids(client, []*SnmpRequest{remPortId}, func(req *SnmpRequest) error {

		localPortNumber, ok := AssumePortNumberOnPos(req, req.GetSplittedOid(), -2)
		if !ok {
			return nil
		}

		oidEnds, ok := AssumeNumberOnPos(req, req.GetSplittedOid(), -1)
		if !ok {
			return nil
		}

		remPortId := req.ParseSnmpResult()
		if remPortId == "" {
			WriteAll("[%s]LLDP-MIB::%s [%s] REQEST: REMOTE PORT ID IS EMPTY! MAYBE UNSUPPORTED OID SCHEME",
				host.Ip, req.name, req.oid)
		}

		remoteLldpPort := host.GetOrCreateRemoteLldpPort(localPortNumber, oidEnds)
		remoteLldpPort.Id = utils.Sanitize(remPortId)

		return nil
	})

	remPortDesc := CraftSnmpRequest(host, "LLDP-MIB::lldpRemPortDesc", ResultTypeString)
	_ = WalkOids(client, []*SnmpRequest{remPortDesc}, func(req *SnmpRequest) error {

		localPortNumber, ok := AssumePortNumberOnPos(req, req.GetSplittedOid(), -2)
		if !ok {
			return nil
		}

		oidEnds, ok := AssumeNumberOnPos(req, req.GetSplittedOid(), -1)
		if !ok {
			return nil
		}

		remPortDesc := req.ParseSnmpResult()

		remoteLldpPort := host.GetOrCreateRemoteLldpPort(localPortNumber, oidEnds)
		remoteLldpPort.Desc = remPortDesc

		return nil
	})

	remSysName := CraftSnmpRequest(host, "LLDP-MIB::lldpRemSysName", ResultTypeString)
	_ = WalkOids(client, []*SnmpRequest{remSysName}, func(req *SnmpRequest) error {

		localPortNumber, ok := AssumePortNumberOnPos(req, req.GetSplittedOid(), -2)
		if !ok {
			return nil
		}

		oidEnds, ok := AssumeNumberOnPos(req, req.GetSplittedOid(), -1)
		if !ok {
			return nil
		}

		remSysName := req.ParseSnmpResult()

		remoteLldpPort := host.GetOrCreateRemoteLldpPort(localPortNumber, oidEnds)
		remoteLldpPort.SysName = remSysName

		return nil
	})

	remSysDesc := CraftSnmpRequest(host, "LLDP-MIB::lldpRemSysDesc", ResultTypeString)
	_ = WalkOids(client, []*SnmpRequest{remSysDesc}, func(req *SnmpRequest) error {

		localPortNumber, ok := AssumePortNumberOnPos(req, req.GetSplittedOid(), -2)
		if !ok {
			return nil
		}

		oidEnds, ok := AssumeNumberOnPos(req, req.GetSplittedOid(), -1)
		if !ok {
			return nil
		}

		remSysDesc := req.ParseSnmpResult()

		remoteLldpPort := host.GetOrCreateRemoteLldpPort(localPortNumber, oidEnds)
		remoteLldpPort.SysDescr = remSysDesc

		return nil
	})

	remChassisId := CraftSnmpRequest(host, "LLDP-MIB::lldpRemChassisId", ResultTypeAuto)
	_ = WalkOids(client, []*SnmpRequest{remChassisId}, func(req *SnmpRequest) error {

		localPortNumber, ok := AssumePortNumberOnPos(req, req.GetSplittedOid(), -2)
		if !ok {
			return nil
		}

		oidEnds, ok := AssumeNumberOnPos(req, req.GetSplittedOid(), -1)
		if !ok {
			return nil
		}

		remChassisId := req.ParseSnmpResult()
		if remChassisId == "" {
			WriteAll("[%s]LLDP-MIB::%s [%s] REQEST: REMOTE PORT CHASSIS ID IS EMPTY! MAYBE UNSUPPORTED OID SCHEME",
				host.Ip, req.name, req.oid)
		}

		remoteLldpPort := host.GetOrCreateRemoteLldpPort(localPortNumber, oidEnds)
		remoteLldpPort.ChassisId = utils.Sanitize(remChassisId)

		return nil
	})

	WriteAll("[%s] LLDP probes DONE!", task.ip)
}
