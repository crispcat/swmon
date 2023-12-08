package main

import (
	goSNMP "swmon_mapper/gosnmp-fixed"
	"time"
)

func CreateSnmpClient(task *NetTask) *goSNMP.GoSNMP {

	version := SNMP_VERSION
	switch task.swargs.SnmpVersion {
	case 1:
		version = goSNMP.Version1
	case 2:
		version = goSNMP.Version2c
	case 3:
		ErrorAll("SNMP Version 3 is not supported. Config:\n%s", task.swargs)
	default:
		ErrorAll("Invalid SNMP version provided. Config:\n%s", task.swargs)
	}

	return &goSNMP.GoSNMP{
		Target:    task.ip.String(),
		Port:      task.swargs.SnmpPort,
		Version:   version,
		Community: task.swargs.SnmpCommunityString,
		Timeout:   time.Duration(task.swargs.SnmpTimeout) * time.Millisecond,
		Retries:   int(task.swargs.SnmpRetries),
	}
}

func SnmpStart(client *goSNMP.GoSNMP) error {

	err := client.Connect()
	if err != nil {
		WriteAll("[%s] Unable to establish SNMP connection with the host: %s", client.Target, err)
		return err
	}

	return nil
}

func SnmpClose(client *goSNMP.GoSNMP) {

	err := client.Conn.Close()
	if err != nil {
		WriteAll("[%s] Dirty close of SNMP connection to the host, error: %s", client.Target, err)
	}
}

func SnmpStartRoutine(task *NetTask, queue *NetTaskQueue) (client *goSNMP.GoSNMP, finalizer func(), err error) {

	client = CreateSnmpClient(task)
	err = SnmpStart(client)
	if err != nil {
		return nil, nil, err
	}

	finalizer = func() {
		queue.DoneOne()
		SnmpClose(client)
	}

	return client, finalizer, nil
}

func GetOids(client *goSNMP.GoSNMP, requests []*SnmpRequest) error {

	oids := make([]string, len(requests))
	for i, request := range requests {
		oids[i] = request.oid
	}

	response, err := client.Get(oids)
	if err != nil {
		WriteAll("[%s] Unable to get SNMP vars from the host: %s", client.Target, err)
		return err
	}

	for i := 0; i < len(oids); i++ {
		req := requests[i]
		req.result = response.Variables[i]
		req.LogSnmpResult()
		req.WiteToHost()
	}

	return nil
}

func WalkOids(client *goSNMP.GoSNMP, requests []*SnmpRequest, callback func(req *SnmpRequest) error) error {

	for i, request := range requests {
		err := client.Walk(request.oid, func(result goSNMP.SnmpPDU) error {
			req := requests[i]
			req.oid = result.Name
			req.result = result
			req.LogSnmpResult()
			req.WiteToHost()

			err := callback(req)
			if err != nil {
				return err
			}

			return nil
		})

		if err != nil {
			WriteAll("[%s] Unable to walk SNMP vars from the host: %s", client.Target, err)
			return err
		}
	}

	return nil
}
