package main

import (
	"flag"
	"math/big"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	shared "swmon_shared"
	"time"
)

const ETC_PATH = shared.ETC

const MAN = `
SWMON_MAPPER

NAME
swmon_mapper OPTION... [OPTION...]

DESCRIPTION

	L2 network topology mapper. Run with root privs.

	REQUIRED

	-n	    Network blocks in CIDR format (192.0.0.1/16). Comma separated.

	-r 	    Root addresses from where fun begins. Comma separated. Must be in provided netblocks.

	-m 	    Path to the NagVis map. Use it to create or update existing NagVis map.

	NOT REQUIRED

	-w 	    Number of parralel workers. Big number increase speed but network stability may lay down due to packet loss.
                    Recomended 500-2000. If your network range is less, set workers count accordingly to your addresses range. 
                    If not provided, this is default befaviour.

	-s 	    SNMP Community string. Default is "public".

	-l 	    Path to logs file. Default logs is swmon_log in ` + ETC_PATH + ` .

	MODES

	-c 	    Use /usr/local/etc/swmon/ config. If config don't exists creates a new one. Mode. Not required.

	-k 	    Use swmon_hosts_model.json generated with previous scans to define targets. No scan will process.
                    Will retrive SNMP data only from hosts already found. -n will be ignored. Mode. Not required.

	-f 	    Forget already found hosts if it become unreachable. Mode. Not required.

	-h --help   Prints that.

EXAMPLE

	Use -c to create default config
      > sudo swmon_mapper -c

	Run with default config
      > sudo swmon_mapper -c

	Process full scan of the subnet with exac 200 workers:
      > sudo swmon_mapper -n 192.168.225.0/24 -s somecomstr -r 192.168.225.201 -w 200

	Remap already found hosts from root and remove unreachable hosts:
      > sudo swmon_mapper -r 192.168.225.33 -k -f

`

// There are command line options only
var (
	Help           bool
	UseConfig      bool
	OnlyKnownHosts bool
)

// Config are args set both from command line and config
var Config SwmonConfig

// SwmonConfig are args set both from command line and config
type SwmonConfig struct {
	LogsPath          string             `yaml:"logs_path"`
	Workers           uint               `yaml:"workers"`
	RootAddr          string             `yaml:"root_addr"`
	NagvisMap         string             `yaml:"nagvis_map"`
	WwwUser           string             `yaml:"www_user"`
	Networks          []SwmonNetworkArgs `yaml:"networks"`
	ForgetHosts       bool               `yaml:"remove_unreachable_hosts"`
	AutoRestartNagios bool               `yaml:"auto_restart_nagios"`
}

// SwmonNetworkArgs are args situable for every separate network block
type SwmonNetworkArgs struct {
	AddrBlocks          string `yaml:"addr_blocks"`
	SnmpCommunityString string `yaml:"snmp_v1_v2c_community_string"`
	SnmpPort            uint16 `yaml:"snmp_port"`
	SnmpVersion         uint8  `yaml:"snmp_version"`
	SnmpTimeout         uint64 `yaml:"snmp_timeout"`
	MsgFlags            string `yaml:"snmp_v3_msg_flags"`
}

func main() {

	startTime := time.Now()

	err := DoLogs(LOG_FILE)
	ParseArgsAndConfig()
	DoneLogs()

	err = DoLogs(Config.LogsPath)
	if err != nil {
		ErrorAll("Unable to open/create logs file %s: %s. Logs will be written in stdout.", Config.LogsPath, err.Error())
	}

	var hostsModel *HostsModel
	var addressesCount *big.Int

	switch OnlyKnownHosts {
	case true:
		hostsModel, addressesCount = ScanKnownHosts()
	default:
		hostsModel, addressesCount = ScanNetwork()
	}

	err = WriteNagiosHostsConfigFile(hostsModel)
	if err != nil {
		WriteAll("Unable to save hosts file. Scaned data will not persist. Err: %s", err.Error())
	}

	DrawMap(hostsModel, Config.NagvisMap)

	err = WriteHostsModel(hostsModel)
	if err != nil {
		WriteAll("Unable to save data file. Scaned data will not persist. Err: %s", err.Error())
	}

	timeElapsed := time.Since(startTime)
	WriteAll("Swmon execution done in %s for %s unique addresses.", timeElapsed, addressesCount)

	if Config.AutoRestartNagios {
		RestartNagios()
		Stdout("Map avaliable on http://localhost/nagvis/frontend/nagvis-js/index.php?mod=Map&act=view&show=%s",
			strings.Split(filepath.Base(Config.NagvisMap), ".")[0])
	}

	DoneLogs()
}

func ParseArgsAndConfig() {

	comandLineNetworkArgs := DEFAULT_NETWORK_ARGS

	flag.BoolVar(&Help, "help", false, "Prints that.")
	flag.BoolVar(&Help, "h", false, "Prints that.")
	flag.BoolVar(&UseConfig, "c", false, "Path to config file. Default config is near the binary with name swmon_config.")
	//flag.BoolVar(&Config.AutoRestartNagios, "a", false, "Auto restart Nagios process when done.")
	flag.StringVar(&comandLineNetworkArgs.AddrBlocks, "n", "", "Network blocks in CIDR format (192.0.0.1/16). Comma separated. Required.")
	flag.StringVar(&Config.RootAddr, "r", "", "Root addresses from where fun begins. Must be in provided netblock. Required.")
	flag.UintVar(&Config.Workers, "w", 0, "Number of parralel workers.")
	flag.StringVar(&Config.LogsPath, "l", LOG_FILE, "Path to logs file. Default logs is here with name swmon_log. Not required.")
	flag.BoolVar(&OnlyKnownHosts, "k", false, "Use swmon_hosts_model.json generated with previous scans to define targets. No network scan will process, only known hosts instead.")
	flag.StringVar(&comandLineNetworkArgs.SnmpCommunityString, "s", SNMP_COMMUNITY, "SNMP Community string. Default is \"public\".")
	flag.StringVar(&Config.NagvisMap, "m", "", "Path to the NagVis map. Use it to gracefully update your existing NagVis map.")
	flag.BoolVar(&Config.ForgetHosts, "f", false, "Forget already finded host if it become unreachable.")
	flag.Parse()

	if Help {
		Stdout(MAN)
		DoneLogs()
		os.Exit(0)
	}

	if UseConfig {
		CreateDefaultConfigIfNExist()
		Config = ParseConfig(CONFIG_PATH)
	}

	flag.Parse()

	if !UseConfig {
		Config.Networks = []SwmonNetworkArgs{comandLineNetworkArgs}
	}

	for i, netw := range Config.Networks {
		if netw.AddrBlocks == "" {
			WriteAll("Configuration error. Address blocks is not provided for network %d.", i)
			Config.Networks = append(Config.Networks[:i], Config.Networks[i+1:]...)
		}
	}

	if len(Config.Networks) == 0 {
		ErrorAll("At least one network block must be configured using config or command line.")
	}
}

type SwmonNetwork struct {
	NetBlocks      []*net.IPNet
	Netwalkers     []*Netwalker
	AddressesCount *big.Int
	Args           SwmonNetworkArgs
}

func ScanNetwork() (*HostsModel, *big.Int) {

	var addrBlocks strings.Builder
	networks := make([]*SwmonNetwork, len(Config.Networks))
	for i, swargs := range Config.Networks {
		addrBlocks.WriteString(swargs.AddrBlocks)
		networks[i] = &SwmonNetwork{}
		networks[i].NetBlocks = ParseNetworkBlocks(swargs)
		networks[i].Args = swargs
	}

	rootIp := ParseRootIp()

	LoadMibs()

	addressesCount := big.NewInt(int64(0))
	for _, nw := range networks {
		nw.Netwalkers = PopulateNetwalkers(nw.NetBlocks)
		nw.AddressesCount = CalculateAddressesCount(nw.Netwalkers)
		addressesCount.Add(addressesCount, nw.AddressesCount)
	}

	SetupWorkersForAddresses(addressesCount)

	WriteAll("Swmon full scan started for blocks %s. It is %d unique addresses. Num workers: %d",
		addrBlocks.String(), addressesCount, Config.Workers)

	hostsModel := CreateHostsModel()
	knownHosts, err := ReadHostsModel()
	if err != nil {
		hostsModel.Import(knownHosts)
	}

	taskQueue := CreateTaskQueue(1024)
	for w := uint(0); w < Config.Workers; w++ {
		go NetWorker(taskQueue, hostsModel)
	}

	for _, network := range networks {
		WalkAddressBlock(network, taskQueue)
	}

	taskQueue.WaitAllTasksCompletesAndClose()
	LinkHosts(hostsModel, rootIp)

	return hostsModel, addressesCount
}

func ScanKnownHosts() (*HostsModel, *big.Int) {

	rootIp := ParseRootIp()

	LoadMibs()

	hosts, err := ReadHostsModel()
	if err != nil {
		return ScanNetwork()
	}

	if Config.Workers == 0 {
		Config.Workers = uint(len(hosts)) * 2
	}

	WriteAll("Swmon monitoring started for hosts file %s! Is is %d unique hosts. Num workers: %d",
		HOSTS_MODEL_FILE, len(hosts), Config.Workers)

	hostsModel := CreateHostsModel()
	hostsModel.Import(hosts)

	taskQueue := CreateTaskQueue(256)
	for w := uint(0); w < Config.Workers; w++ {
		go NetWorker(taskQueue, hostsModel)
	}

	for _, host := range hosts {
		taskQueue.Enqueue(NetTask{ip: host.Ip, swargs: host.NetworkArgs, method: Ping})
	}

	taskQueue.WaitAllTasksCompletesAndClose()
	LinkHosts(hostsModel, rootIp)

	return hostsModel, big.NewInt(int64(len(hosts)))
}

func ParseNetworkBlocks(swargs SwmonNetworkArgs) []*net.IPNet {

	addrBlocks := strings.Split(swargs.AddrBlocks, ",")
	for i, addrBlock := range addrBlocks {
		addrBlocks[i] = strings.Trim(addrBlock, " ")
	}

	netBlocks := make([]*net.IPNet, len(addrBlocks))
	for i, addrBlock := range addrBlocks {
		var err error
		_, netBlocks[i], err = net.ParseCIDR(addrBlock)
		if err != nil {
			ErrorAll("Invalid network address block provided: %s. See --help.", err.Error())
		}
	}

	return netBlocks
}

func ParseRootIp() net.IP {

	if Config.RootAddr == "" {
		ErrorAll("Root address is not provided. See --help.")
	}

	rootIp := net.ParseIP(Config.RootAddr)
	if rootIp == nil {
		ErrorAll("Provided root %s is not IP address.", Config.RootAddr)
	}

	rootIp = rootIp[len(rootIp)-4:]
	if rootIp == nil {
		ErrorAll("Invalid root address provided: %s. See --help.", Config.RootAddr)
	}

	return rootIp
}

func PopulateNetwalkers(netBlocks []*net.IPNet) []*Netwalker {

	nw := CreateNetwalker(netBlocks[0])
	netwalkers := []*Netwalker{nw}
	for i := 1; i < len(netBlocks); i++ {
		if !nw.IncludeIfCan(netBlocks[i]) {
			netwalkers = append(netwalkers, CreateNetwalker(netBlocks[i]))
		}
	}

	return netwalkers
}

func CalculateAddressesCount(netwalkers []*Netwalker) *big.Int {

	count := big.NewInt(0)
	for _, nw := range netwalkers {
		count.Add(count, big.NewInt(int64(nw.AddressRange())))
	}

	return count
}

func SetupWorkersForAddresses(addressesCount *big.Int) {

	if Config.Workers == 0 {
		if addressesCount.Cmp(big.NewInt(1024)) >= 0 {
			Config.Workers = 1024
		} else {
			Config.Workers = uint(addressesCount.Uint64())
		}
	}
}

func WalkAddressBlock(network *SwmonNetwork, taskQueue *NetTaskQueue) {

	for _, nw := range network.Netwalkers {
		nw.Current = nw.Start
		for {
			addr := nw.CurrentAddress()
			taskQueue.Enqueue(NetTask{ip: addr, swargs: network.Args, method: Ping})

			if !nw.HaveNextAddress() {
				break
			}

			err := nw.GoNextAddress()
			if err != nil {
				ErrorAll(err.Error())
			}
		}
	}
}

func RestartNagios() {

	err := exec.Command("/bin/sh", "-c", "sudo systemctl restart nagios.service").Run()
	if err != nil {
		WriteAll("Unable to restart Nagios service: %s", err)
	} else {
		WriteAll("Nagios service restarted!")
	}
}

func GetExecPath() string {

	execPath, err := os.Executable()
	if err != nil {
		return ""
	}

	return execPath
}
