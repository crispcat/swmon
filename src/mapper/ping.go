package main

import (
	"fmt"
	"github.com/go-ping/ping"
	"net"
	"strings"
	"time"
)

func Ping(task *NetTask, queue *NetTaskQueue, hostsModel *HostsModel) {

	defer queue.DoneOne()
	Verbose("[%s] Send ping...", task.ip)

	pinger, err := ping.NewPinger("127.0.0.1") // use dummy address to pass check
	if err != nil {
		WriteAll("[%s] Unable to create pinger for: %s", task.ip, err)
		return
	}

	pinger.SetIPAddr(&net.IPAddr{IP: task.ip}) // set address directly from raw bytes
	pinger.SetPrivileged(true)
	pinger.Count = 3
	pinger.Timeout = 3 * time.Second

	err = pinger.Run()
	if err != nil {
		WriteAll("[%s] Pinger encounted an error trying to probe: %s", task.ip, err)
		return
	}

	if pinger.PacketsRecv != 0 {
		var output strings.Builder
		output.WriteString(fmt.Sprintf("[%s] Reachable!\n", task.ip))
		stats := pinger.Statistics()
		output.WriteString(fmt.Sprintf("[%s] Ping statistics: ", task.ip))
		output.WriteString(fmt.Sprintf("%d packets transmitted, %d packets received, %v packet loss ",
			stats.PacketsSent, stats.PacketsRecv, stats.PacketLoss))
		output.WriteString(fmt.Sprintf("round-trip min/avg/max/stddev = %v/%v/%v/%v",
			stats.MinRtt, stats.AvgRtt, stats.MaxRtt, stats.StdDevRtt))

		WriteAll(output.String())

		host := hostsModel.GetOrCreate(task.ip)
		host.NetworkArgs = task.swargs

	} else if host := hostsModel.Get(task.ip); Config.ForgetUnreachable && host != nil {
		var output strings.Builder
		output.WriteString(fmt.Sprintf("[%s] Unreachable!\n", task.ip))
		output.WriteString(fmt.Sprintf("[%s] Will be deleted from map!\n", task.ip))
		WriteAll(output.String())
		hostsModel.Delete(task.ip)
	}
}
