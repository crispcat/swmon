package main

import (
	"flag"
	"math/big"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// There are command line options only
var (
	Help             bool
	UseDefaultConfig bool
	OnlyKnownHosts   bool
	IsVerbose        bool
	ConfigPath       string
	ForgetAllHosts   bool
)

// Config are args both from command line and config
var Config SwmonConfig

type SwmonConfig struct {
	LogsPath          string             `yaml:"logs_path"`
	Workers           uint               `yaml:"workers"`
	RootAddr          string             `yaml:"root_addr"`
	NagvisMap         string             `yaml:"nagvis_map"`
	WwwUser           string             `yaml:"www_user"`
	Networks          []SwmonNetworkArgs `yaml:"networks"`
	ForgetUnreachable bool               `yaml:"remove_unreachable_hosts"`
	PostExecCommand   string             `yaml:"post_execution_command"`
}

// SwmonNetworkArgs are args situable for every separate network block
type SwmonNetworkArgs struct {
	AddrBlocks          string `yaml:"addr_blocks"`
	SnmpCommunityString string `yaml:"snmp_community_string"`
	SnmpPort            uint16 `yaml:"snmp_port"`
	SnmpVersion         uint8  `yaml:"snmp_version"`
	SnmpTimeout         uint64 `yaml:"snmp_timeout"`
	SnmpRetries         uint8  `yaml:"snmp_retries"`
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

	if Config.PostExecCommand != "" {
		PostExecCommand()
	}

	WriteAll("Map avaliable on http://localhost/nagvis/frontend/nagvis-js/index.php?mod=Map&act=view&show=%s",
		strings.Split(filepath.Base(Config.NagvisMap), ".")[0])

	DoneLogs()
}

func ParseArgsAndConfig() {

	comandLineNetworkArgs := DEFAULT_NETWORK_ARGS

	flag.BoolVar(&Help, "help", false, H_DESCR)
	flag.BoolVar(&Help, "h", false, H_DESCR)
	flag.BoolVar(&UseDefaultConfig, "c", false, C_DESCR)
	flag.StringVar(&ConfigPath, "conf", "", CONF_DESCR)
	flag.StringVar(&comandLineNetworkArgs.AddrBlocks, "n", "", N_DESCR)
	flag.StringVar(&Config.RootAddr, "r", "", R_DESCR)
	flag.UintVar(&Config.Workers, "w", 0, W_DESCR)
	flag.StringVar(&Config.LogsPath, "l", LOG_FILE, L_DESCR)
	flag.BoolVar(&OnlyKnownHosts, "k", false, K_DESCR)
	flag.StringVar(&comandLineNetworkArgs.SnmpCommunityString, "s", SNMP_COMMUNITY, S_DESCR)
	flag.StringVar(&Config.NagvisMap, "m", "", M_DESCR)
	flag.BoolVar(&Config.ForgetUnreachable, "f", false, F_DESCR)
	flag.BoolVar(&ForgetAllHosts, "ff", false, FF_DESCR)
	flag.BoolVar(&IsVerbose, "v", false, V_DESCR)
	flag.Parse()

	if Help {
		Stdout(MAN)
		DoneLogs()
		os.Exit(0)
	}

	UseCustomConfig := ConfigPath != ""

	if UseCustomConfig {
		CreateConfigIfNExist(ConfigPath)
		Config = ParseConfig(ConfigPath)

	} else if UseDefaultConfig {
		CreateConfigIfNExist(DEFAULT_CONFIG_PATH)
		Config = ParseConfig(DEFAULT_CONFIG_PATH)
	}

	// give a priority to command line arguments
	flag.Parse()

	if !UseDefaultConfig && !UseCustomConfig {
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
	if err != nil || ForgetAllHosts {
		WriteAll("Unable to read model file %s: %s", HOSTS_MODEL_FILE, err)
	} else {
		hostsModel.Import(knownHosts)
	}

	taskQueue := CreateTaskQueue()
	for w := uint(0); w < Config.Workers; w++ {
		go NetWorker(taskQueue, hostsModel)
	}

	WriteAll("Sending ICMP...")
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
	if err != nil || ForgetAllHosts {
		return ScanNetwork()
	}

	if Config.Workers == 0 {
		Config.Workers = uint(len(hosts)) * 2
	}

	WriteAll("Swmon monitoring started for hosts file %s! Is is %d unique hosts. Num workers: %d",
		HOSTS_MODEL_FILE, len(hosts), Config.Workers)

	hostsModel := CreateHostsModel()
	hostsModel.Import(hosts)

	taskQueue := CreateTaskQueue()
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

func PostExecCommand() {

	err := exec.Command("/bin/sh", "-c", Config.PostExecCommand).Run()
	if err != nil {
		WriteAll("Unable to do post execution: %s", err)
	} else {
		WriteAll("Post execution succeed!")
	}
}

func GetExecPath() string {

	execPath, err := os.Executable()
	if err != nil {
		return ""
	}

	return execPath
}
