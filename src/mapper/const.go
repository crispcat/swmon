package main

import (
	"regexp"
	shared "swmon_shared"
)

//-----------------------------------------CONFIG---------------------------------------------------------------------//

var NEW_CONFIG = SwmonConfig{
	LogsPath:          LOG_FILE,
	Workers:           0,
	RootAddr:          "",
	WwwUser:           "www-data",
	NagvisMap:         "/usr/local/nagvis/etc/maps/swmon-static.cfg",
	Networks:          []SwmonNetworkArgs{DEFAULT_NETWORK_ARGS},
	ForgetUnreachable: false,
	PostExecCommand:   "sudo systemctl restart nagios",
}

var DEFAULT_NETWORK_ARGS = SwmonNetworkArgs{
	SnmpPort:            SNMP_PORT,
	SnmpCommunityString: SNMP_COMMUNITY,
	SnmpVersion:         2,
	SnmpTimeout:         15_000,
	SnmpRetries:         5,
}

//---------------------------------------------SNMP-------------------------------------------------------------------//

const SNMP_PORT = uint16(161)
const SNMP_VERSION = Version2c
const SNMP_COMMUNITY = "public"

//--------------------------------------------FILES-------------------------------------------------------------------//

const DEFAULT_CONFIG_PATH = shared.ETC + "swmon_config.yaml"

const OS_DIR_PERMISSIONS = 0755
const OS_FILE_PERMISSIONS_STRICT = 0660
const OS_FILE_PERMISSIONS_R = 0664
const BACKUP_ROOT = shared.ETC + "backup/"

const HOSTS_MODEL_FILE = shared.ETC + "swmon_hosts_model.json"
const HOSTS_MODEL_FILE_BACKUP_PREFIX = BACKUP_ROOT + "swmon_hosts_model_"

const HOSTS_CONFIG_FILE = shared.ETC + "swmon_nagios_hosts.cfg"
const HOSTS_CONFIG_FILE_BACKUP_PREFIX = BACKUP_ROOT + "swmon_nagios_hosts_"

const LOG_FILE = shared.ETC + "swmon_log"

//----------------------------------------------STRINGS-----------------------------------------------------------------//

const hexDigit = "0123456789ABCDEF"

var HardwareAddressRegex = regexp.MustCompile("^((([0-9A-F]{2}[ :-]){5})|(([0-9A-F]{2}[ :-]){7})|(([0-9A-F]{2}[ :-]){19}))([0-9A-F]{2})$")
var TEMPLATE_REPLACE_REGEX = regexp.MustCompile(`#\w+#`)

const SERVICE_GROUP_FORMAT_STRING = "%s:Port%s__%s:Port%s"

//--------------------------------------------------------------------------------------------------------------------//
