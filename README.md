# SWMON
### Simple autodiscovering LLDP network mapper for Nagios and Nagvis with STP monitoring support

![img.png](img_1.png)

## How It Works
Swmon scans network blocks with ping. Then, try to gather system information and LLDP links from alive hosts via SNMP. Finally, it work to "connect all the ends" and build a logical topology of the network.

### Output
As a result of the scan swmon will generate three files:
- `swmon_nagios_hosts.cfg` - [Nagios](https://github.com/NagiosEnterprises/nagioscore)
  config with hosts, child/parent relations, and services to monitor the network.
  You can configure Nagios to include this.
- [Nagvis](https://github.com/NagVis/nagvis) static map config written to path you decide.
- `swmon_hosts_model.json` - state of the network during last scan.

There are nothing else. Swmon just runs and generate three config files.   

All futher monitoring on Nagios behalf.  

Visualization - on Nagvis.

### Network devices requirements

- SNMPv2 enabled on the devices (SNMPv3 is not supported yet)
- Known community string
- IPv4 addresses
- LLDP data available by SNMP
- STP data available by SNMP

### What I need to see it?

- [Nagios](https://github.com/NagiosEnterprises/nagioscore)
- [Nagios Plugins](https://github.com/nagios-plugins/nagios-plugins)
- [Nagvis*](https://github.com/NagVis/nagvis)
- Some [Nagvis backend](http://docs.nagvis.org/1.9/en_US/backends.html)   

I use [ndoutils](https://github.com/NagiosEnterprises/ndoutils) + MySQL 8.* as Nagvis backend to monitor swmon generated topologies, but you free to use something else.

*Use version > 1.9.34 if you want to use ndoutils + MySQL 8.* (you can install it from my submodule in /deps. I fixed some minor MySQL 8.* incompatibility)

## Installation

### From sources
Go 1.19 is required to build the sources.
```bash
make install clean
```

### From build

Download the latest build, then:

```bash
 chmod +x install.sh
 ./install.sh
```

The installation script will copy necessary binaries to `/usr/local/bin`  
and program files to `/usr/local/etc/swmon`

## Usage

### Setup
First, you need to create a default config

```
 # sudo swmon_mapper -c
 Default config created!
```

```yaml
logs_path: /usr/local/etc/swmon/swmon_log
workers: 0
root_addr: ""
nagvis_map: /usr/local/nagvis/etc/maps/swmon-static.cfg
www_user: www-data
networks:
   - addr_blocks: ""
     snmp_community_string: public
     snmp_port: 161
     snmp_version: 2
     snmp_timeout: 15000
remove_unreachable_hosts: false
post_execution_command: sudo systemctl restart nagios
```

It is a bit of settings. But required are `root_addr` and `nagvis_map` with at
least one network block. See detailed config below.

```yaml
logs_path: /usr/local/etc/swmon/swmon_log
workers: 0 # is auto, see --help for more details
root_addr: "192.168.14.1" # will have no parents and be the N-gen parent for all other hosts
nagvis_map: /usr/local/nagvis/etc/maps/swmon-static.cfg # nagvis map will created or updated
www_user: www-data # your www data user
networks: # can be a few
   - addr_blocks: "192.168.14.0/22,192.168.10.0/24" # CIDR format only
     snmp_community_string: somecomstring
     snmp_port: 161
     snmp_version: 2 # can be 1
     snmp_timeout: 15000

   - addr_blocks: "172.17.22.0/24"
     snmp_community_string: somecomstring
     snmp_port: 161
     snmp_version: 2
     snmp_timeout: 15000

remove_unreachable_hosts: false # make sense while updating existing maps
post_execution_command: sudo systemctl restart nagios
```

Or use command line instead:

```bash
sudo swmon_mapper -n 192.168.225.0/24 -s public -r 192.168.225.201 -w 256
sudo swmon_mapper --help
```

### First run

When configured.

```bash
sudo swmon_mapper -c
```
```
Testing LLDP MIBs persistence...
LLDP-MIB::lldpLocPortId: 1.0.8802.1.1.2.1.3.7.1.3
LLDP-MIB::lldpRemPortIdSubtype: 1.0.8802.1.1.2.1.4.1.1.6
LLDP-MIB::lldpRemPortId: 1.0.8802.1.1.2.1.4.1.1.7
LLDP-MIB::lldpRemPortDesc: 1.0.8802.1.1.2.1.4.1.1.8
LLDP-MIB::lldpRemSysName: 1.0.8802.1.1.2.1.4.1.1.9
LLDP-MIB::lldpRemSysDesc: 1.0.8802.1.1.2.1.4.1.1.10
LLDP-MIB::lldpRemChassisIdSubtype: 1.0.8802.1.1.2.1.4.1.1.4
LLDP-MIB::lldpRemChassisId: 1.0.8802.1.1.2.1.4.1.1.5
LLDP-MIB::lldpRemSysCapSupporte: 1.0.8802.1.1.2.1.4.1.1.11
LLDP-MIB::lldpRemSysCapEnabled: 1.0.8802.1.1.2.1.4.1.1.12
Swmon full scan started for blocks 192.168.14.0/22,192.168.10.0/24,172.17.22.0/24. It is 1536 unique addresses. Num workers: 1024
[192.168.12.22] Send ping...
[192.168.12.8] Send ping...
[192.168.12.1] Send ping...
[192.168.15.255] Send ping...
[192.168.15.28] Send ping...
[192.168.12.14] Send ping...
...
```
You will see a lot of output.  
But interesting things are in the last lines

```
...
MAPPER: CREATING NAGVIS MAP /usr/local/nagvis/etc/maps/swmon-static.cfg...
MAPPER: CREATING NEW NAGVIS MAP ON PATH /usr/local/nagvis/etc/maps/swmon-static.cfg
Swmon execution done in 6.264937363s for 1536 unique addresses.
Map available on http://localhost/nagvis/frontend/nagvis-js/index.php?mod=Map&act=view&show=swmon-static

```
Navigate to the `/usr/local/etc/swmon` directory:
```
drwxr-xr-x 5 root root    4096 ./
drwxr-xr-x 3 root root    4096 ../
drw-r--r-- 2 root root    4096 backup/
drwxr-xr-x 2 root root    4096 maps/
drwxr-xr-x 3 root root    4096 mibs/
-rw-r----- 1 root root     366 swmon_config.yaml
-rw-r--r-- 1 root root  183116 swmon_hosts_model.json
-rw-r--r-- 1 root root 1723078 swmon_log
-rw-r--r-- 1 root root   14537 swmon_nagios_hosts.cfg
```

You will see Nagios config `swmon_nagios_hosts.cfg`.  
Include it to your Nagios config.
```
sudo bash -c "echo "cfg_file=$(realpath swmon_nagios_hosts.cfg)" >> /usr/local/nagios/etc/nagios.cfg"
sudo systemctl restart nagios
```

Follow the link at the end of swmon output. All hosts will be located same coordinates in top the
left corner. Click Edit Map -> Lock/Unlock All and drag it as you want.

Don't worry about orange "UNKNOWN" services we will fix it soon.

## STP Monitoring

After the first scan, you will end with orange connections telling us that the status of the services is
**'UNKNOWN'**.

![img_3.png](img_3.png)
![img_4.png](img_4.png)

It is a predicted behavior because we don't tell swmon which STP protocol devices are using
and how to retrieve STP statuses.  
You need to edit `/usr/local/etc/swmon/mibs/oids/stp_dev_oids.yaml`.

```yaml
- device_match_regex: DGS-1210-52/ME
  target_oid: SNMPv2-SMI::enterprises.171.10.76.29.1.6.2.1.12
  value_map: mstp_statuses

- device_match_regex: .*
  target_oid: MSTP-MIB::swMSTPMstPortStatus
  value_map: mstp_statuses
```

It's MSTP monitoring setup I made for the test network (4 DLink switches). Two of them supports
[MSTP-MIB](http://www.circitor.fr/Mibs/Html/M/MSTP-MIB.php) and the`MSTP-MIB::swMSTPMstPortStatus`
is the root OID for port statuses.

You can use the already shipped tool `stp_mib_scanner` to parse STP module from the device `sysORtable`

```
stp_mib_scanner -c some_community_str -h host_ip -w
```
And see output like follows:
```
Device name: DGS-3000-28L
Device description: DGS-3000-28L Gigabit Ethernet Switch
STP module name parsed from device table: swMSTPMIB
STP module OID parsed from device table: .1.3.6.1.4.1.171.12.15
Searching for supported user MIB module. The name of the module may be swMSTPMIB.mib...
Module swMSTPMIB.mib not found in user MIBs.
STP module OID .1.3.6.1.4.1.171.12.15 from device table is valid OID
ROOT STP OID IS: .1.3.6.1.4.1.171.12.15
Press ENTER to list module values...
```
Now, you can use OID to search for the MIB module. I use [oidref](https://oidref.com/).

![img.png](img.png)

Download the MIB, place it in `/usr/local/etc/swmon/mibs` directory, and name as the module name in the device `sysORtable`.
```
sudo mv ~/Downloads/dlkMSTP.mib /usr/local/etc/swmon/mibs/swMSTPMIB.mib
stp_mib_scanner -c somecomstr -h 192.14.1.18 -w
```
```
...
Module swMSTPMIB.mib found!
Searching for root node...
Root node is swMSTPMIB
STP module OID parsed from user MIB module swMSTPMIB.mib is 1.3.6.1.4.1.171.12.15
STP module OID .1.3.6.1.4.1.171.12.15 from device table is valid OID
ROOT STP OID IS: 1.3.6.1.4.1.171.12.15
Press ENTER to list module values...
```
Press ENTER and you will see port statuses oids somewhere in output.
```
...
MSTP-MIB::swMSTPMstPortStatus.20.0 = INTEGER: disabled(2)
MSTP-MIB::swMSTPMstPortStatus.21.0 = INTEGER: disabled(2)
MSTP-MIB::swMSTPMstPortStatus.22.0 = INTEGER: disabled(2)
MSTP-MIB::swMSTPMstPortStatus.23.0 = INTEGER: forwarding(5)
MSTP-MIB::swMSTPMstPortStatus.24.0 = INTEGER: forwarding(5)
MSTP-MIB::swMSTPMstPortStatus.25.0 = INTEGER: disabled(2)
MSTP-MIB::swMSTPMstPortStatus.26.0 = INTEGER: disabled(2)
...
```
Add `MSTP-MIB::swMSTPMstPortStatus` to you config.
I matched all device models with that OID.
```
- device_match_regex: .*
  target_oid: MSTP-MIB::swMSTPMstPortStatus
```

### If there are no suitable MIB modules you can targeting raw OID. For example:
```
Device name: DGS-1210-52/ME
Device description: DGS-1210-52/ME/A1
STP module name parsed from device table: dlinkMSTP
STP module OID parsed from device table: .1.3.6.1.4.1.171.10.76.29.1.6
Searching for supported user MIB module. Name of module may be dlinkMSTP.mib...
Module dlinkMSTP.mib not found in user MIBs.
STP module OID .1.3.6.1.4.1.171.10.76.29.1.6 from device table is valid OID
ROOT STP OID IS: .1.3.6.1.4.1.171.10.76.29.1.6
Press ENTER to list module values...
```
There are no public record for oid .1.3.6.1.4.1.171.10.76.29.1.6  
List values and try to find port statuses somewhere in the module.
```
...
DLINK-ID-REC-MIB::dlink-products.76.29.1.6.2.1.12.44 = INTEGER: 1
DLINK-ID-REC-MIB::dlink-products.76.29.1.6.2.1.12.45 = INTEGER: 1
DLINK-ID-REC-MIB::dlink-products.76.29.1.6.2.1.12.46 = INTEGER: 1
DLINK-ID-REC-MIB::dlink-products.76.29.1.6.2.1.12.47 = INTEGER: 5
DLINK-ID-REC-MIB::dlink-products.76.29.1.6.2.1.12.48 = INTEGER: 5
DLINK-ID-REC-MIB::dlink-products.76.29.1.6.2.1.12.49 = INTEGER: 1
DLINK-ID-REC-MIB::dlink-products.76.29.1.6.2.1.12.50 = INTEGER: 1
...
```
We knew that ports 47 and 48 are linked. According to MSTP values map value 5 is "forwarding".
```
mstp_statuses:
    1: "Unknown: other(1)"
    2: "Unknown: disabled(2)"
    3: "Warning: discarding(3)"
    4: "Unknown: learning(4)"
    5: "OK: forwarding(5)"
    6: "Critical: broken(6)"
    7: "Unknown: no-stp-enabled(7)"
    8: "Critical: err-disabled(8)"
```
Yes, 1 stands for "other", but anyway
OID `DLINK-ID-REC-MIB::dlink-products.76.29.1.6.2.1.12` looks like the best candidate.

Add it to the config
```
- device_match_regex: DGS-1210-52/ME
  target_oid: DLINK-ID-REC-MIB::dlink-products.76.29.1.6.2.1.12
  value_map: mstp_statuses

- device_match_regex: .*
  target_oid: MSTP-MIB::swMSTPMstPortStatus
  value_map: mstp_statuses
```

Do not forget to clear swmon cache every time you need to change target OIDs.
```
sudo rm /tmp/swmon_*
```

![img_5.png](img_5.png)

You can also add your own values maps in `/usr/local/etc/swmon/mibs/oids/value_maps.yaml` file and
monitor whatever protocol you want.

## Some usage tips

- **[Set](https://support.nagios.com/forum/viewtopic.php?f=7&t=60813) Nagios `interval_length` to `1` for "realtime" monitoring.**
- **Swmon will not entirely rewrite your map. Swmon will not delete any host already on the map.
  If the map exists it will only add new hosts found, and update data for the old hosts (links etc).
  So, you can safely customize map and then run swmon scans. Use -f flag if you want to delete some old hosts not presented in the network anymore.**
- **Contact me if you run into some issues. I don't have access to any big enough network to test all possible cases.**

## Licence

This is free software under GNU GENERAL PUBLIC LICENSE. See LICENCE for more details.

## Thanks

Big thanks to the developers of [GoSNMP](https://github.com/gosnmp/gosnmp) library for such a relief.
