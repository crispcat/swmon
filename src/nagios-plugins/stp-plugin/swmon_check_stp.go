package main

//--------------------------------------------------------------------------------------------------------------------//

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gosnmp/gosnmp"
	"os"
	"strings"
	plugin "swmon_shared"
)

//--------------------------------------------------------------------------------------------------------------------//

func main() {

	flag.StringVar(&plugin.Args.Host, "h", "", "-h 127.0.0.1 (HOST IP ADDRESS)")
	flag.StringVar(&plugin.Args.SnmpCommunityString, "c", "public", "-c public (COMMUNITY STRING)")
	flag.UintVar(&plugin.Args.SnmpPort, "p", uint(161), "-p 161 (SNMP PORT NUMBER)")
	flag.UintVar(&plugin.Args.SnmpVersion, "v", uint(2), "-v 2 (SNMP VERSION)")
	flag.BoolVar(&plugin.Args.IsDebug, "d", false, "-d (Enables DEBUG mode)")
	flag.UintVar(&plugin.Args.Iface, "i", 0, "-i (INTERFACE/PORT NUMBER TO CHECK)")
	flag.Parse()

	if plugin.Args.Iface == 0 {
		plugin.Exit_Unknown("Interface number (PORT) is not provided.")
	}

	if plugin.Args.Host == "" {
		plugin.Exit_Unknown("Host is not provided.")
	}

	client := plugin.CreateSnmpClient(plugin.Args)
	if err := client.Connect(); err != nil {
		plugin.Exit_Unknown("Unable to connect to SNMP agent h:%s p:%s v:%d c:%s",
			plugin.Args.Host, plugin.Args.SnmpPort, plugin.Args.SnmpVersion, plugin.Args.SnmpCommunityString)
	}

	defer func() {
		_ = client.Conn.Close()
	}()

	plugin.LoadMibModules()

	var oid string
	var oids []string
	var valueMap plugin.ValueMap
	var valueMapName string

	// read cahce
	cacheFileName := fmt.Sprintf(plugin.TMP+"swmon_%s:%d.cache", plugin.Args.Host, plugin.Args.Iface)
	bytes, err := os.ReadFile(cacheFileName)
	if err != nil {

		plugin.Debug("Cache file %s not found. Will try to match device with record in oid table.", cacheFileName)
		var stpPortStatusOid string
		stpPortStatusOid, valueMap, valueMapName = MatchStpPortStatusOid(client)
		oid = fmt.Sprintf("%s.%d", plugin.MibsGetOid(stpPortStatusOid), plugin.Args.Iface)
		oids = []string{oid, oid + ".0"}

	} else {

		plugin.Debug("Found cache file %s. Loading...", cacheFileName)
		cache := strings.Split(string(bytes), "=")
		err := json.Unmarshal([]byte(cache[1]), &valueMap)
		if err != nil {
			plugin.Exit_Unknown("Unable to unmarshal value map from cache: %s", err)
		}
		oid = cache[0]
		oids = []string{oid}
	}

	plugin.Debug("STP port status oid: %s", oid)

	var portStatus string
	var portStatusOid string
	portStatusRes, err := client.Get(oids)
	for _, v := range portStatusRes.Variables {

		plugin.Debug("Probing %s", v.Name)
		portStatus = plugin.ParseAsUint(v.Value)
		if portStatus != "" {
			portStatusOid = v.Name
			break
		}

		plugin.Debug("UInt value %s is empty. Trying to get string %v...", v.Name, v.Value)
		portStatus = plugin.ParseString(v.Value)
		portStatus = plugin.Sanitize(portStatus)
		if portStatus != "" {
			portStatusOid = v.Name
			break
		}

		plugin.Debug("String value %s is empty. Go next.", v.Name)
	}

	if portStatus == "" {
		plugin.Exit_Unknown("No port status value finded on oid tree %s", oid)
	}

	// write cache
	valueMapCache, err := json.Marshal(valueMap)
	if err != nil {
		plugin.Exit_Unknown("Unable to marshal value map to cache: %s", err)
	}

	err = os.WriteFile(cacheFileName, append([]byte(portStatusOid+"="), valueMapCache...), os.FileMode(0644).Perm())
	if err != nil {
		plugin.Debug("Unable to write cache file %s: %s", cacheFileName, err.Error())
	}

	portStatus = plugin.Sanitize(portStatus)
	result, ok := valueMap[portStatus]
	if !ok {
		plugin.Exit_Unknown("Unable to match value %s in value map %s: %s", portStatus, valueMapName, err)
	}

	plugin.Exit(result)
}

func MatchStpPortStatusOid(client *gosnmp.GoSNMP) (string, map[string]string, string) {

	// get device model
	oids := []string{plugin.MibsGetOid("sysDescr.0")}
	sysDescrRes, err := client.Get(oids)
	if err != nil {
		plugin.Exit_Unknown("Unable to get device description: %s", err.Error())
	}

	// match device model with regex pattern in oids config
	sysDesrc := plugin.ParseString(sysDescrRes.Variables[0].Value)
	oidMap, err := plugin.ParseOidMap(plugin.DEV_STP_OID_CONFIG_FILE)
	if err != nil {
		plugin.Exit_Unknown("Unable to parse device OID config file %s: %s", plugin.DEV_STP_OID_CONFIG_FILE, err.Error())
	}

	portStatusesOid, err := oidMap.GetOid(sysDesrc)
	if err != nil {
		plugin.Exit_Unknown("Unable to match device record for device %s in %s", sysDesrc, plugin.DEV_STP_OID_CONFIG_FILE)
	}

	valueMaps, err := plugin.ParseValueMaps(plugin.VALUE_MAPS_FILE)
	if err != nil {
		plugin.Exit_Unknown("Unable to load values map %s: %s", plugin.VALUE_MAPS_FILE, err)
	}

	valueMap, ok := valueMaps[portStatusesOid.ValueMapName]
	if !ok {
		plugin.Exit_Unknown("Unable to find value map %s: %s", portStatusesOid.ValueMapName, err)
	}

	return portStatusesOid.TargetOid, valueMap, portStatusesOid.ValueMapName
}

//--------------------------------------------------------------------------------------------------------------------//
