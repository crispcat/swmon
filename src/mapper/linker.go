package main

import (
	"fmt"
	"net"
	"sync"
)

func LinkHosts(model *HostsModel, rootIp net.IP) {

	WriteAll("[Linker] Linking hosts...")

	var rootHost = model.Get(rootIp)
	if rootHost == nil {
		ErrorAll("[LINKER] Root host %s isn`t discovered during the scan. FATAL. Unable to link hosts from the root.", rootIp)
	}

	model.IndexHosts()

	//hostsCount := model.HostsCount()

	workers := uint32(1) //+ hostsCount/10
	hostsChan := make(chan *Host, 256)
	wg := &sync.WaitGroup{}

	for i := uint32(0); i < workers; i++ {
		go LinkWorker(model, hostsChan, wg)
	}

	wg.Add(1)
	hostsChan <- rootHost

	wg.Wait()
	close(hostsChan)
}

func LinkWorker(model *HostsModel, hostsChan chan *Host, wg *sync.WaitGroup) {

	for current := range hostsChan {

		for _, lldpPort := range current.LldpPorts {

			if len(lldpPort.RemotePorts) == 0 {
				continue
			}

			if len(lldpPort.RemotePorts) > 1 {
				WriteAll("[LINKER] LINK AGREGATION DETECTED %s --> %s --> %s",
					current, lldpPort, lldpPort.RemotePorts)
			}

			for _, remotePort := range lldpPort.RemotePorts {

				var searchReq string

				if HardwareAddressRegex.MatchString(remotePort.Id) {
					searchReq = remotePort.Id

				} else if HardwareAddressRegex.MatchString(remotePort.ChassisId) {
					searchReq = remotePort.ChassisId

				} else {
					searchReq = remotePort.SysName
				}

				hostOwnable, ok := model.Index[searchReq]

				if ok {
					host := hostOwnable.GetOwner()
					WriteAll("[LINKER] HOST FOUND ON PATH %s --> %s --> %s --> %s",
						current, lldpPort, remotePort, host)

					linkDescr := fmt.Sprintf("%d --> %s", lldpPort.SubOid, host.Name)
					current.LinksDescr[linkDescr] = true

					if conn, ok := model.ConnectionGet(current, host); ok {
						// once we met already discovered path, we update original conn toPort number
						if !conn.done {
							conn.toPort = lldpPort.SubOid
							conn.done = true
							model.ConnectionSet(conn)
						}

						WriteAll("[LINKER] SKIPPING ALREDY DISCOVERED PATH %s --> %s --> %s --> %s",
							current, lldpPort, remotePort, host)
						continue
					}

					conn := Connection{from: current.Id, to: host.Id, frPort: lldpPort.SubOid}
					model.ConnectionSet(conn)
					host.AddParent(current)
					remotePort.owner = host
					hostsChan <- host
					wg.Add(1)

				} else {
					WriteAll("[LINKER] ENDPOINT HOST NOT INDEXED %s --> %s --> %s --> %s",
						current, lldpPort, remotePort, "NULL")
				}
			}
		}

		wg.Done()
	}
}
