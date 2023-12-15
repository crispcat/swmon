# SWMON
### Simple autodiscovering network mapper for Nagios and Nagvis.

<a href="https://imgbb.com/"><img src="https://i.ibb.co/M9tFTvh/img-1.png" alt="img-1" border="0"></a>

## How It Works
Swmon scans network blocks with ping. Then, try to gather system information from alive hosts via SNMP.

### Output
As a result of the scan swmon will generate three files:
- `swmon_nagios_hosts.cfg` - [Nagios](https://github.com/NagiosEnterprises/nagioscore)
  config with hosts, and services to monitor the network.
  You can configure Nagios to include this.
- [Nagvis](https://github.com/NagVis/nagvis) static map config written to path you decide.
- `swmon_hosts_model.json` - state of the network during last scan.

### Network devices requirements

- SNMPv2 enabled on the devices
- Known community string
- IPv4 addresses

### What I need to see it?

- [Nagios](https://github.com/NagiosEnterprises/nagioscore)

### Optional (for maps):
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
 Config created on path /usr/local/etc/swmon/swmon_config.yaml!
```

**Use -conf instead of -c to specify exac config path*

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
sudo swmon_mapper -help
```

### First run

When configured.

```bash
sudo swmon_mapper -c
```
**Use -conf instead of -c to specify exac config path*
```
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

## Some usage tips

<a href="https://ibb.co/ZzqLKtj"><img src="https://i.ibb.co/d48W61T/img-2.png" alt="img-2" border="0"></a>

- **[Set](https://support.nagios.com/forum/viewtopic.php?f=7&t=60813) Nagios `interval_length` to `1` if you need "realtime" monitoring for a while**
- **Swmon will not entirely rewrite your map. Swmon will not delete any host already on the map.
  If the map exists it will only add a new hosts found and update data for the old hosts.
  So, you can safely customize map and then run swmon scans. Use the -f flag if you want to delete some old hosts not presented in a network anymore.**
- **Contact me if you run into some issues. I don't have access to any big enough network to test all possible cases.**

## Licence

This is free software under GNU GENERAL PUBLIC LICENSE. See LICENCE for more details.

## Thanks

Big thanks to the developers of [GoSNMP](https://github.com/gosnmp/gosnmp) library for such a relief.
