package main

import shared "swmon_shared"

const ETC_PATH = shared.ETC

const N_DESCR = "Network blocks in CIDR format (192.0.0.1/16). Comma separated."

const R_DESCR = "Root addresses from where fun begins. Comma separated. Must be in provided netblocks."

const M_DESCR = "Path to the NagVis map. Use it to create or update existing NagVis map."

const W_DESCR = `Number of parralel workers. Big number increase speed but network stability may lay down due to packet loss.
                    Recomended 500-2000. If your network range is less, set workers count accordingly to your addresses range.
                    If not provided, this is default befaviour.`

const S_DESCR = "SNMP Community string. Default is \"public\"."

const L_DESCR = "Path to logs file. Default logs is swmon_log in " + ETC_PATH

const C_DESCR = "Use " + DEFAULT_CONFIG_PATH + " config. If config don't exists creates a new one. Mode. Not required."

const CONF_DESCR = "Use given config. If config don't exists creates a new one. Mode. Not required."

const K_DESCR = `Use swmon_hosts_model.json generated with previous scans to define targets. No scan will process.
                    Will retrive SNMP data only from hosts already found. -n will be ignored. Mode. Not required.`

const F_DESCR = "Forget already found hosts if it become unreachable. Mode. Not required."

const FF_DESCR = "Forget all host already found. Mode. Not required."

const V_DESCR = "Be verbose."

const H_DESCR = "Prints that."

const MAN = `
SWMON_MAPPER

NAME
swmon_mapper OPTION... [OPTION...]

DESCRIPTION

	Autodiscovering LLDP network mapper for Nagios and Nagvis. Run it with root privs.

	REQUIRED

	-n	    ` + N_DESCR + `

	-r 	    ` + R_DESCR + `

	-m 	    ` + M_DESCR + `

	NOT REQUIRED

	-w 	    ` + W_DESCR + `

	-s 	    ` + S_DESCR + `

	-l 	    ` + L_DESCR + `

	MODES

	-c 	    ` + C_DESCR + `

	-conf       ` + CONF_DESCR + `

	-k 	    ` + K_DESCR + `

	-f 	    ` + F_DESCR + `

	-ff         ` + FF_DESCR + `

	-v 	    ` + V_DESCR + `

	-h --help   ` + H_DESCR + `

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
