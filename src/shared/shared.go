package shared

import (
	"errors"
	"fmt"
	"github.com/gosnmp/gosnmp"
	"github.com/hallidave/mibtool/smi"
	"gopkg.in/yaml.v3"
	"os"
	"regexp"
	"strings"
	"time"
)

const ETC = "/usr/local/etc/swmon/"
const TMP = "/tmp/"
const MIBS_ROOT = ETC + "mibs"
const OIDS_ROOT = ETC + "mibs/oids/"
const DEV_STP_OID_CONFIG_FILE = OIDS_ROOT + "stp_oid_map.yaml"
const VALUE_MAPS_FILE = OIDS_ROOT + "value_maps.yaml"

var MIB = smi.NewMIB(MIBS_ROOT)
var ERR_BREAK = errors.New("b")

//--------------------------------------------------------------------------------------------------------------------//

type PluginArgs struct {
	Host                string
	SnmpCommunityString string
	SnmpPort            uint
	SnmpVersion         uint
	IsDebug             bool
	Iface               uint
}

var Args = PluginArgs{
	Host:                "",
	SnmpCommunityString: "public",
	SnmpPort:            161,
	SnmpVersion:         2,
	IsDebug:             false,
}

//--------------------------------------------------------------------------------------------------------------------//

func Exit(exitStatusMessage string) {

	exitStatus := strings.Split(exitStatusMessage, ":")[0]

	switch exitStatus {

	case "OK":
		Exit_OK(exitStatusMessage)

	case "Warning":
		Exit_Warning(exitStatusMessage)

	case "Critical":
		Exit_Critical(exitStatusMessage)

	case "Unknown":
		Exit_Unknown(exitStatusMessage)

	default:
		Exit_Unknown("Invalid status string %s!", exitStatus)
	}
}

func Exit_OK(format string, v ...any) {
	fmt.Printf(format, v...)
	os.Exit(0)
}

func Exit_Warning(format string, v ...any) {
	fmt.Printf(format, v...)
	os.Exit(1)
}

func Exit_Critical(format string, v ...any) {
	fmt.Printf(format, v...)
	os.Exit(2)
}

func Exit_Unknown(format string, v ...any) {
	fmt.Printf(format, v...)
	os.Exit(3)
}

func Debug(format string, v ...any) {
	if Args.IsDebug {
		fmt.Printf(format+"\n", v...)
	}
}

//--------------------------------------------------------------------------------------------------------------------//

const SNMP_TIMEOUT = 15_000 // 15 sec

func CreateSnmpClient(args PluginArgs) *gosnmp.GoSNMP {

	var version gosnmp.SnmpVersion
	switch args.SnmpVersion {
	case 1:
		version = gosnmp.Version1
	case 2:
		version = gosnmp.Version2c
	case 3:
		Exit_Unknown("SNMP Version 3 is not supported.")
	default:
		Exit_Unknown("Invalid SNMP version provided.")
	}

	return &gosnmp.GoSNMP{
		Target:    args.Host,
		Port:      uint16(args.SnmpPort),
		Version:   version,
		MsgFlags:  gosnmp.NoAuthNoPriv,
		Community: args.SnmpCommunityString,
		Timeout:   time.Duration(SNMP_TIMEOUT) * time.Millisecond,
		MaxOids:   100,
	}
}

func SnmpVerToString(snmpVer uint) (string, error) {

	switch snmpVer {
	case 1:
		return "1", nil
	case 2:
		return "2c", nil
	default:
		return "", fmt.Errorf("invalid SNMP version %d priovided", snmpVer)
	}
}

func MibsGetOid(name string) string {

	oid, err := MIB.OID(name)
	if err != nil {
		Exit_Unknown("FATAL! Unable to get OID for %s", name)
	}

	return oid.String()
}

//--------------------------------------------------------------------------------------------------------------------//

type DeviceOid struct {
	DeviceMatchRegex string `yaml:"device_match_regex"`
	TargetOid        string `yaml:"target_oid"`
	ValueMapName     string `yaml:"value_map"`
}

type OidMap []DeviceOid

type ValueMap map[string]string

type ValueMaps map[string]ValueMap

func (oidMap OidMap) GetOid(deviceModel string) (DeviceOid, error) {

	for _, record := range oidMap {

		regex, err := regexp.Compile(record.DeviceMatchRegex)
		if err != nil {
			return DeviceOid{}, err
		}

		match := regex.MatchString(deviceModel)
		if match {
			return record, nil
		}
	}

	return DeviceOid{}, errors.New("no matched device record found")
}

func ParseOidMap(filename string) (OidMap, error) {

	bytes, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	oidMap := OidMap{}
	err = yaml.Unmarshal(bytes, &oidMap)

	if err != nil {
		return nil, err
	}

	return oidMap, nil
}

func ParseValueMaps(filename string) (ValueMaps, error) {

	bytes, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	valueMaps := ValueMaps{}
	err = yaml.Unmarshal(bytes, &valueMaps)

	if err != nil {
		return nil, err
	}

	return valueMaps, nil
}

func LoadMibModules() {

	if err := MIB.LoadModules(); err != nil {
		Exit_Unknown("Unable to load MIB modules on path ./mibs: %s", err)
	}

	if Args.IsDebug {
		if len(MIB.Modules) == 0 {
			Debug("No MIB modules loaded.")
		}
		for _, m := range MIB.Modules {
			Debug("Loaded MIB module %s", m.Name)
		}
	}
}

//--------------------------------------------------------------------------------------------------------------------//
