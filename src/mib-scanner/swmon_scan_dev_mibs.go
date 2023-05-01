package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/gosnmp/gosnmp"
	"github.com/hallidave/mibtool/smi"
	"gopkg.in/yaml.v3"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	plugin "swmon_shared"
)

var STP_REGEX = regexp.MustCompile(`(.)*((([RMVrmv])*[Ss][Tt][Pp])|([Rr]*[Pp][Vv][Ss][Tt]))(.)*`)

const SYS_NAME = "sysName.0"
const SYS_DESCRIPTION = "sysDescr.0"
const SYS_OR_DESCRIPTION = "sysORDescr"
const SYS_OR_ID = "sysORID"
const VALID_ROOT_OID_LENGHT = 7

func Error(format string, v ...any) {
	fmt.Printf(format+"\n", v...)
	os.Exit(1)
}

func Verb(format string, v ...any) {
	fmt.Printf(format+"\n", v...)
}

func main() {

	var oidsCreateModule string
	var walkFindedModule bool

	flag.StringVar(&plugin.Args.Host, "h", "", "-h 127.0.0.1 (HOST IP ADDRESS)")
	flag.StringVar(&plugin.Args.SnmpCommunityString, "c", "public", "-c public (COMMUNITY STRING)")
	flag.UintVar(&plugin.Args.SnmpPort, "p", uint(161), "-p 161 (SNMP PORT NUMBER)")
	flag.UintVar(&plugin.Args.SnmpVersion, "v", uint(2), "-v 2 (SNMP VERSION)")
	flag.StringVar(&oidsCreateModule, "o", "", "-o stp (CREATES STP OIDS CONFIG)")
	flag.BoolVar(&walkFindedModule, "w", false, "-w (WALKS FINDED MODULE WITH SNMPWALK)")
	flag.Parse()

	if oidsCreateModule != "" {
		JustCreateConfig(oidsCreateModule)
	}

	if plugin.Args.Host == "" {
		Error("Host is not provided.")
	}

	client := plugin.CreateSnmpClient(plugin.Args)
	if err := client.Connect(); err != nil {
		Error("Unable to connect to SNMP agent h:%s p:%s v:%d c:%s",
			plugin.Args.Host, plugin.Args.SnmpPort, plugin.Args.SnmpVersion, plugin.Args.SnmpCommunityString)
	}

	defer func() {
		_ = client.Conn.Close()
	}()

	if err := plugin.MIB.LoadModules(); err != nil {
		Error("Unable to load MIB modules on path ./mibs: %s", err)
	}

	rootOid := FindStpRoot(client)
	Verb("ROOT STP OID IS: %s", rootOid)

	if walkFindedModule {
		WalkFindedModule(rootOid)
	}
}

func WalkFindedModule(oid string) {

	plugin.PressEnterTo("list module values")

	snmpVer, err := plugin.SnmpVerToString(plugin.Args.SnmpVersion)
	if err != nil {
		Error(err.Error())
	}

	walkCmd := fmt.Sprintf("snmpwalk -v%s -M +'./mibs' -m +ALL -c %s %s -Ci %s",
		snmpVer, plugin.Args.SnmpCommunityString, plugin.Args.Host, oid)

	cmd := exec.Command("/bin/sh", "-c", walkCmd)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()

	if err != nil {
		Error(err.Error())
	}
}

func JustCreateConfig(oidsCreateModule string) {

	newConfig := plugin.OidMap{{}, {}}

	bytes, err := yaml.Marshal(newConfig)
	if err != nil {
		Error("Unable to marshal empty config file: %s", err.Error())
	}

	var path string
	switch oidsCreateModule {
	case "stp":
		path = plugin.DEV_STP_OID_CONFIG_FILE
	default:
		Error("Module %s is not implemented", oidsCreateModule)
	}

	err = os.WriteFile(path, bytes, os.FileMode(0644))
	if err != nil {
		Error("Unable to create new config file for module %s: %s", oidsCreateModule, err.Error())
	}

	Verb("Config for module %s created on path %s", oidsCreateModule, path)
	os.Exit(0)
}

func FindStpRoot(client *gosnmp.GoSNMP) string {

	portSearchRoots := make([]string, 0, 2)

	stpModuleRootOidFromTable, stpModuleName := GetStpModuleOidFromDeviceTable(client)
	if stpModuleName == "" {
		Error("No avalible STP modules found in the device table.")
	}

	Verb("STP module name parsed from device table: %s", stpModuleName)
	Verb("STP module OID parsed from device table: %s", stpModuleRootOidFromTable)

	stpModuleOidFromMibModule := GetStpModuleOidFromMibModule(stpModuleName)

	if stpModuleOidFromMibModule != "" {
		Verb("STP module OID parsed from user MIB module %s.mib is %s", stpModuleName, stpModuleOidFromMibModule)
		portSearchRoots = append(portSearchRoots, stpModuleOidFromMibModule)
	}

	oidIsValid := len(strings.Split(stpModuleRootOidFromTable, ".")) >= VALID_ROOT_OID_LENGHT
	if stpModuleRootOidFromTable != "" && oidIsValid {
		Verb("STP module OID %s from device table is valid OID", stpModuleRootOidFromTable)
		portSearchRoots = append(portSearchRoots, stpModuleRootOidFromTable)
	}

	if len(portSearchRoots) == 0 {
		Error("No avalible STP modules found. But device support it! "+
			"Download needed mib and place it as %s.mib in mibs folder.", stpModuleName)
	}

	return portSearchRoots[0]
}

func GetStpModuleOidFromDeviceTable(client *gosnmp.GoSNMP) (string, string) {

	oids := []string{plugin.MibsGetOid(SYS_NAME), plugin.MibsGetOid(SYS_DESCRIPTION)}
	res, err := client.Get(oids)
	if err != nil {
		Error("Unable to get SNMP OID %s: %s", SYS_NAME, err)
	}

	Verb("Device name: %s", plugin.ParseString(res.Variables[0].Value))
	Verb("Device description: %s", plugin.ParseString(res.Variables[1].Value))

	var moduleOid string
	var moduleName string
	oid := plugin.MibsGetOid(SYS_OR_DESCRIPTION)
	err = client.Walk(oid, func(res gosnmp.SnmpPDU) error {
		moduleName = STP_REGEX.FindString(plugin.ParseString(res.Value))
		if moduleName != "" {
			moduleOid = res.Name
			return plugin.ERR_BREAK
		}
		return nil
	})

	if err != nil && !errors.Is(err, plugin.ERR_BREAK) {
		Error("Unable to walk SNMP OID %s: %s", SYS_OR_DESCRIPTION, err)
	}

	if moduleName == "" {
		Error("[R/M]STP or [R]PVST support record not found in the device table.")
	}

	moduleSysIndex, err := plugin.GetUint32(strings.Split(moduleOid, "."), -1)
	if err != nil {
		Error("Oid pos -1 on %s is not a number", moduleOid)
	}

	moduleIdOid := fmt.Sprintf("%s.%d", plugin.MibsGetOid(SYS_OR_ID), moduleSysIndex)
	sysOrModuleRootOidRes, err := client.Get([]string{moduleIdOid})
	if err != nil {
		Error("Unable to get module ID on OID %s: %s", moduleIdOid, err.Error())
	}

	moduleRootOid := sysOrModuleRootOidRes.Variables[0].Value.(string)
	return moduleRootOid, moduleName
}

func GetStpModuleOidFromMibModule(stpModuleName string) string {

	Verb("Searching for supported user MIB module. Name of module may be %s.mib...", stpModuleName)
	var module *smi.Module
	for _, m := range plugin.MIB.Modules {
		fileName := strings.Split(filepath.Base(m.File), ".")[0]
		if fileName == stpModuleName {
			module = m
			break
		}
	}

	if module == nil {
		Verb("Module %s.mib not found in user MIBs.", stpModuleName)
		return ""
	}

	Verb("Module %s.mib found!", stpModuleName)
	Verb("Searching for root node...")

	var rootNode *smi.Node
	for _, n := range module.Nodes {
		if n.Type == smi.NodeModuleID {
			rootNode = &n
			break
		}
	}

	if rootNode == nil {
		Verb("Module %s.mib contains no ModuleID Node", stpModuleName)
		return ""
	}

	Verb("Root node is %s", rootNode.Label)

	moduleRootOid, err := plugin.MIB.OID(rootNode.Label)
	if err != nil {
		Verb("Cannot get OID for root node entity %s", rootNode.Label)
		return ""
	}

	return moduleRootOid.String()
}
