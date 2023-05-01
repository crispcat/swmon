package main

import utils "swmon_shared"

func SNMP_Ifs(task *NetTask, queue *NetTaskQueue, hostsMap *HostsModel) {

	client, finalizer, err := SnmpStartRoutine(task, queue)
	if err != nil {
		return
	}

	defer finalizer()

	WriteAll("[%s] Starting Interfaces probes...", task.ip)

	host := hostsMap.GetOrCreate(task.ip)

	// supress errors to retrieve all data we can

	ifIndex := CraftSnmpRequest(host, "ifIndex", ResultTypeAuto)
	_ = WalkOids(client, []*SnmpRequest{ifIndex}, func(req *SnmpRequest) error {

		portNumber, ok := AssumePortNumberOnPos(req, req.GetSplittedOid(), -1)
		if !ok {
			return nil
		}

		ifIndex := req.ParseSnmpResult()
		if ifIndex == "" {
			WriteAll("[%s] SNMP GET %s [%s] REQEST: INTERFACE INDEX NOT PARSED! MAYBE UNSUPPORTED OID SCHEME",
				host.Ip, req.name, req.oid)
		}

		ifs := host.GetOrCreateInterface(portNumber)
		ifs.Index = utils.Sanitize(ifIndex)

		return nil
	})

	ifDescr := CraftSnmpRequest(host, "ifDescr", ResultTypeString)
	_ = WalkOids(client, []*SnmpRequest{ifDescr}, func(req *SnmpRequest) error {

		portNumber, ok := AssumePortNumberOnPos(req, req.GetSplittedOid(), -1)
		if !ok {
			return nil
		}

		ifDescr := req.ParseSnmpResult()
		if ifDescr == "" {
			WriteAll("[%s] SNMP GET %s [%s] REQEST: INTERFACE DESCR NOT PARSED! MAYBE UNSUPPORTED OID SCHEME",
				host.Ip, req.name, req.oid)
		}

		ifs := host.GetOrCreateInterface(portNumber)
		ifs.Descr = ifDescr

		return nil
	})

	ifPhysAddress := CraftSnmpRequest(host, "ifPhysAddress", ResultTypeAuto)
	_ = WalkOids(client, []*SnmpRequest{ifPhysAddress}, func(req *SnmpRequest) error {

		portNumber, ok := AssumePortNumberOnPos(req, req.GetSplittedOid(), -1)
		if !ok {
			return nil
		}

		ifPhysAddress := req.ParseSnmpResult()
		if ifPhysAddress == "" {
			WriteAll("[%s] SNMP GET %s [%s] REQEST: INTERFACE PHYS ADDRESS NOT PARSED! MAYBE UNSUPPORTED OID SCHEME",
				host.Ip, req.name, req.oid)
		}

		ifs := host.GetOrCreateInterface(portNumber)
		ifs.MAC = ifPhysAddress

		return nil
	})

	ifOperStatus := CraftSnmpRequest(host, "ifOperStatus", ResultTypeInteger)
	_ = WalkOids(client, []*SnmpRequest{ifOperStatus}, func(req *SnmpRequest) error {

		portNumber, ok := AssumePortNumberOnPos(req, req.GetSplittedOid(), -1)
		if !ok {
			return nil
		}

		ifOperStatus := req.ParseSnmpResult()
		if ifOperStatus == "" {
			WriteAll("[%s] SNMP GET %s [%s] REQEST: INTERFACE OPER STATUS NOT PARSED! MAYBE UNSUPPORTED OID SCHEME",
				host.Ip, req.name, req.oid)
		}

		ifs := host.GetOrCreateInterface(portNumber)

		switch utils.Sanitize(ifOperStatus) {
		case "1":
			// up
			ifs.Operational = true
		case "2":
			// down
			ifs.Operational = false
		default:
			WriteAll("[%s] SNMP GET %s [%s] REQEST: INTERFACE OPER STATUS NOT VALID (%s) ! MAYBE UNSUPPORTED OID SCHEME",
				host.Ip, req.name, req.oid, ifOperStatus)
		}

		return nil
	})

	WriteAll("[%s] Interfaces probes DONE!", task.ip)

	queue.Enqueue(NetTask{ip: task.ip, swargs: task.swargs, method: SNMP_GetLLDP})
}
