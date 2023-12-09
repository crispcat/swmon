package main

import (
	"strconv"
)

func SNMP_SysNameDescr(task *NetTask, queue *NetTaskQueue, hostsMap *HostsModel) {

	client, finalizer, err := SnmpStartRoutine(task, queue)
	if err != nil {
		return
	}

	defer finalizer()

	WriteAll("[%s] Starting SNMP System Name/Description probe...\n", task.ip)

	host := hostsMap.GetOrCreate(task.ip)
	sysNameReq := CraftSnmpRequest(host, "sysName.0", ResultTypeString)
	sysDescReq := CraftSnmpRequest(host, "sysDescr.0", ResultTypeString)
	sysLocReq := CraftSnmpRequest(host, "sysLocation.0", ResultTypeString)

	err = GetOids(client, []*SnmpRequest{sysNameReq, sysDescReq, sysLocReq})
	if err != nil {
		host.WriteToConfig = true
		host.HaveSnmp = false
		return
	}

	desRes := sysDescReq.ParseSnmpResult()
	if desRes == "" || desRes == "0" {
		WriteAll("[%s] SNMP System Name/Description probe result [%s][%s]: %s. IS ZERO OR EMPTY "+
			"MAYBE A MISMATCHED OID!\n",
			task.ip, sysDescReq.name, sysDescReq.oid, desRes)
	}

	nameRes := sysNameReq.ParseSnmpResult()
	if nameRes == "" || nameRes == "0" {
		WriteAll("[%s] SNMP System Name/Description probe result [%s][%s]: %s. IS ZERO OR EMPTY "+
			"MAYBE A MISMATCHED OID OR NAME NOT SET!\n",
			task.ip, sysNameReq.name, sysNameReq.oid, nameRes)
	}

	locRes := sysLocReq.ParseSnmpResult()
	if locRes == "" || locRes == "0" {
		WriteAll("[%s] SNMP System Location probe result [%s][%s]: %s. IS ZERO OR EMPTY "+
			"MAYBE A MISMATCHED OID OR NAME NOT SET!\n",
			task.ip, sysNameReq.name, sysNameReq.oid, nameRes)
	}

	host.SetUniqueName(sanitizeString(nameRes))
	host.WriteToConfig = true
	host.HaveSnmp = true
	host.Description = sanitizeString(desRes)
	host.Location = sanitizeString(locRes)

	WriteAll("[%s] System Name/Description: %s / %s\n", task.ip, nameRes, desRes)
	WriteAll("[%s] System Location: %s\n", task.ip, locRes)

	queue.Enqueue(NetTask{ip: task.ip, swargs: task.swargs, method: SNMP_IfNumber})
}

func SNMP_IfNumber(task *NetTask, queue *NetTaskQueue, hostsMap *HostsModel) {

	client, finalizer, err := SnmpStartRoutine(task, queue)
	if err != nil {
		return
	}

	defer finalizer()

	WriteAll("[%s] Starting Interfaces Count probe...", task.ip)

	host := hostsMap.GetOrCreate(task.ip)
	request := CraftSnmpRequest(host, "ifNumber.0", ResultTypeInteger)

	err = GetOids(client, []*SnmpRequest{request})
	if err != nil {
		return
	}

	ifsCount, err := strconv.ParseUint(request.ParseSnmpResult(), 10, 8)
	if err != nil {
		WriteAll("[%s] SNMP Interfaces Count probe result [%s][%s]: %s. NOT A NUMBER "+
			"MAYBE A MISSMATCHED OID. NOTHING WILL BE WRITTEN TO HOST!\n",
			task.ip, request.name, request.oid, request.result)
		return
	}

	host.IfsCount = uint32(ifsCount)
	if ifsCount > 0 {
		queue.Enqueue(NetTask{ip: task.ip, swargs: task.swargs, method: SNMP_Ifs})
	}

	WriteAll("[%s] Interfaces Count: %d\n", task.ip, ifsCount)
}
