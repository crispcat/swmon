package main

import (
	"encoding/binary"
	"fmt"
	"net"
	"sort"
	"strings"
	"sync"
)

type HostsModel struct {
	Map   sync.Map
	Conn  sync.Map
	Index HostsIndex
}

type Host struct {
	Id            uint32
	Ip            net.IP
	Name          string
	Description   string
	WriteToConfig bool
	WriteToMap    bool
	IfsCount      uint32
	Interfaces    map[uint32]*Interface
	LldpPorts     map[uint32]*LocalLldpPort
	Oids          map[string]string
	Parents       []string
	LinksDescr    map[string]bool
	NetworkArgs   SwmonNetworkArgs
	MapId         string
}

type Interface struct {
	SubOid      uint32
	Index       string
	Descr       string
	MAC         string
	Operational bool

	owner *Host
}

type LocalLldpPort struct {
	SubOid      uint32
	Id          string
	RemotePorts map[uint32]*RemoteLldpPort
}

type RemoteLldpPort struct {
	SubOid    uint32
	Id        string
	Desc      string
	ChassisId string
	SysName   string
	SysDescr  string
	OwnerIp   net.IP

	owner *Host
}

type Connection struct {
	from   uint32
	frPort uint32
	to     uint32
	toPort uint32
	done   bool
}

type HostsIndex map[string]HostOwnable

type HostOwnable interface {
	GetOwner() *Host
}

type HostNameIndex struct {
	Name  string
	owner *Host
}

func CreateHostsModel() *HostsModel {
	return &HostsModel{Map: sync.Map{}, Conn: sync.Map{}, Index: HostsIndex{}}
}

func (hostsModel *HostsModel) Get(ip net.IP) *Host {

	id := binary.BigEndian.Uint32(ip)
	r, exist := hostsModel.Map.Load(id)
	if !exist {
		return nil
	}

	return r.(*Host)
}

func (hostsModel *HostsModel) GetOrCreate(ip net.IP) *Host {

	id := binary.BigEndian.Uint32(ip)
	r, ok := hostsModel.Map.Load(id)

	var host *Host
	if !ok {
		host = CreateHost(ip)
		host.Id = id
	} else {
		host = r.(*Host)
	}

	hostsModel.Map.Store(host.Id, host)
	return host
}

func (hostsModel *HostsModel) Delete(ip net.IP) {

	id := binary.BigEndian.Uint32(ip)
	hostsModel.Map.Delete(id)
}

func (hostsModel *HostsModel) ConnectionGet(from *Host, to *Host) (Connection, bool) {

	key := from.Id ^ to.Id
	conn, ok := hostsModel.Conn.Load(key)
	if !ok {
		return Connection{}, ok
	}
	return conn.(Connection), ok
}

func (hostsModel *HostsModel) ConnectionSet(conn Connection) {

	key := conn.from ^ conn.to
	hostsModel.Conn.Store(key, conn)
}

func (hostsModel *HostsModel) IndexHosts() {

	hostsModel.Map.Range(func(key any, value any) bool {

		host := value.(*Host)

		if host.Name != "" {
			hostsModel.Index[host.Name] = &HostNameIndex{Name: host.Name, owner: host}
		}

		for _, ifs := range host.Interfaces {
			if HardwareAddressRegex.MatchString(ifs.MAC) {
				hostsModel.Index[ifs.MAC] = ifs
			}
		}

		return true
	})
}

func (hostsModel *HostsModel) HostsCount() uint32 {

	var hostsCount uint32
	hostsModel.Map.Range(func(_ any, _ any) bool {
		hostsCount++
		return true
	})

	return hostsCount
}

func (hostsModel *HostsModel) Export() map[uint32]*Host {

	hosts := map[uint32]*Host{}
	hostsModel.Map.Range(func(key, value any) bool {
		hosts[key.(uint32)] = value.(*Host)
		return true
	})

	return hosts
}

func (hostsModel *HostsModel) Import(hostsMap map[uint32]*Host) {

	for id, host := range hostsMap {
		host.OnDeserialize()
		hostsModel.Map.Store(id, host)
	}

	hostsModel.Map.Range(func(_, value any) bool {
		host := value.(*Host)
		for _, locLldpPort := range host.LldpPorts {
			for _, remLldpPort := range locLldpPort.RemotePorts {
				if len(remLldpPort.OwnerIp) == 0 {
					continue
				}
				remLldpPort.OwnerIp = remLldpPort.OwnerIp[len(host.Ip)-4:]
				remLldpPort.owner = hostsModel.Get(remLldpPort.OwnerIp)
			}
		}
		return true
	})
}

func (host *Host) ClearInterfaces() {
	host.Interfaces = map[uint32]*Interface{}
}

func (host *Host) ClearLldp() {
	host.LldpPorts = map[uint32]*LocalLldpPort{}
}

func (host *Host) ClearParents() {
	host.Parents = []string{}
}

func (host *Host) ClearLinkDescriptions() {
	host.LinksDescr = map[string]bool{}
}

func (host *Host) String() string {
	return fmt.Sprintf("[%s]:[%s]:[%s]", host.Ip, host.Name, host.Description)
}

func (ifs *Interface) String() string {
	return fmt.Sprintf("[%d]:[%s]:[%s]:[%t]", ifs.SubOid, ifs.Index, ifs.MAC, ifs.Operational)
}

func (llp *LocalLldpPort) String() string {
	return fmt.Sprintf("[%d]:[%s]", llp.SubOid, llp.Id)
}

func (rlp *RemoteLldpPort) String() string {
	return fmt.Sprintf("[%d]:[%s]:[%s]:[%s]:[%s]:[%s]", rlp.SubOid, rlp.Id, rlp.Desc, rlp.ChassisId, rlp.SysName, rlp.SysDescr)
}

func (hni *HostNameIndex) GetOwner() *Host {

	return hni.owner
}

func (ifs *Interface) GetOwner() *Host {

	return ifs.owner
}

func (rlp *RemoteLldpPort) GetOwner() *Host {

	return rlp.owner
}

func CreateHost(ip net.IP) *Host {

	return &Host{
		Ip:         ip,
		Interfaces: map[uint32]*Interface{},
		LldpPorts:  map[uint32]*LocalLldpPort{},
		Oids:       map[string]string{},
		Parents:    []string{},
		LinksDescr: map[string]bool{},
	}
}

func (host *Host) SetUniqueName(name string) {

	if name == "" {
		host.Name = host.Ip.String()
	} else {
		host.Name = name
	}
}

func sanitizeString(s string) string {
	var replacer = strings.NewReplacer(
		"\r\n", "",
		"\r", "",
		"\n", "",
		"\v", "",
		"\f", "",
		"\u0085", "",
		"\u2028", "",
		"\u2029", "",
	)
	return replacer.Replace(s)
}

func (host *Host) AddParent(parent *Host) {

	host.Parents = append(host.Parents, parent.GetUniqueName())
}

func (host *Host) GetOrCreateInterface(subOid uint32) *Interface {

	ifs, exist := host.Interfaces[subOid]
	if !exist {
		ifs = &Interface{SubOid: subOid}
		ifs.owner = host
		host.Interfaces[subOid] = ifs
	}

	return ifs
}

func (host *Host) GetOrCreateLocalLldpPort(subOid uint32) *LocalLldpPort {

	port, exist := host.LldpPorts[subOid]
	if !exist {
		port = &LocalLldpPort{SubOid: subOid, RemotePorts: map[uint32]*RemoteLldpPort{}}
		host.LldpPorts[subOid] = port
	}

	return port
}

func (host *Host) GetOrCreateRemoteLldpPort(localSubOid uint32, subOid uint32) *RemoteLldpPort {

	localPort := host.GetOrCreateLocalLldpPort(localSubOid)
	remotePort, exist := localPort.RemotePorts[subOid]
	if !exist {
		remotePort = &RemoteLldpPort{SubOid: subOid}
		localPort.RemotePorts[subOid] = remotePort
	}

	return remotePort
}

func (host *Host) OnDeserialize() {

	host.Ip = host.Ip[len(host.Ip)-4:]
	for _, ifs := range host.Interfaces {
		ifs.owner = host
	}
}

func (host *Host) DefaultFormatter(fieldName string) string {

	switch fieldName {

	case "Name":
		return host.GetUniqueName()

	case "Ip":
		return host.Ip.String()

	case "Parents":
		return host.GetParentsString()

	case "Alias":
		return host.GetUniqueAlias()

	case "Notes":
		return host.Report("<b>", "</b><br>") + host.LinksString("<b>", "</b><br>")

	case "Community":
		return host.NetworkArgs.SnmpCommunityString

	default:
		return ""
	}
}

func (host *Host) GetUniqueName() string {

	return host.Name
}

func (host *Host) GetUniqueAlias() string {

	alias := ""
	if host.Name != "" {
		alias += host.Name
	} else {
		alias += host.Ip.String()
	}
	if host.Description != "" {
		alias += ":" + strings.ReplaceAll(host.Description, " ", "_")
	}

	return alias
}

func (host *Host) GetParentsString() string {

	if len(host.Parents) > 0 {
		return strings.Join(host.Parents, ",")

	} else {
		return "localhost"
	}
}

func (host *Host) Report(pre string, post string) string {

	var sb strings.Builder
	sb.WriteString(pre)
	sb.WriteString(fmt.Sprintf("Name: %s", host.Name))
	sb.WriteString(post + pre)
	sb.WriteString(fmt.Sprintf("Description: %s", host.Description))
	sb.WriteString(post)

	return sb.String()
}

func (host *Host) LinksString(pre string, post string) string {

	i := 0
	linksDescr := make([]string, len(host.LinksDescr))
	for ld := range host.LinksDescr {
		linksDescr[i] = ld
		i++
	}

	sort.Strings(linksDescr)

	var sb strings.Builder
	sb.WriteString(pre)
	sb.WriteString(strings.Join(linksDescr, post+pre))
	sb.WriteString(post)

	return sb.String()
}
