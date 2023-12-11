package main

const NAGIOS_CONFIG_HEADER = `
define hostgroup {
    hostgroup_name          switches
    alias                   Network Switches
}
	
define command {
    command_name    check_stp
    command_line    swmon_check_stp -h $ARG1$ -c $ARG2$ -i $ARG3$
}

`

const TEMPLATE_NAGIOS_SWITCH = `
define host {
    use                     generic-switch
    host_name               #Name#
    alias                   #Alias#
    parents                 #Parents#
    address                 #Ip#
    notes                   #Notes#
    hostgroups              switches
    icon_image              switch.gif
    statusmap_image         switch.gd2
}

`

const TEMPLATE_NAGIOS_PING_SERVICE = `
define service {
    use                     generic-service
    host_name               #Name#
    service_description     PING
    check_command           check_ping!200.0,20%!600.0,60%
    check_interval          10
    retry_interval          5
}

`

const TEMPLATE_NAGIOS_UPTIME_SERVICE = `
define service {
    use                     generic-service
    host_name               #Name#
    service_description     Uptime
    check_command           check_snmp!-C #Community# -o sysUpTime.0 -r 1 -t 120 -4 -P 2c
    check_interval          30
    retry_interval          15
}

`

const TEMPLATE_NAGIOS_PORT_LINK_STATUS_SERVICE = `
define service {
   use                     generic-service
   host_name               #Name#
   service_description     #Port# Port Link Status
   check_command           check_snmp!-C #Community# -o ifOperStatus.#Port# -r 1 -t 120 -4 -P 2c
   check_interval          10
   retry_interval          5
}

`

const TEMPLATE_NAGIOS_CONNECTION_SERVICE_GROUP = `
define servicegroup {
    servicegroup_name  #ServiceGroup#
    alias 	       #ServiceGroup#
}

define service {
    use                     generic-service
    host_name               localhost
    servicegroups           #ServiceGroup#
    service_description     #Name1#:Port#Port1# Link Status
    check_command           check_snmp!-H #Ip1# -C #Community1# -o ifOperStatus.#Port1# -r 3 -t 45
    check_interval          10
    retry_interval          5
}

define service {
    use                     generic-service
    host_name               localhost
    servicegroups           #ServiceGroup#
    service_description     #Name2#:Port#Port2# Link Status
    check_command           check_snmp!-H #Ip2# -C #Community2# -o ifOperStatus.#Port2# -r 3 -t 45
    check_interval          10
    retry_interval          5
}

define service {
    use                     generic-service
    host_name               localhost
    servicegroups           #ServiceGroup#
    service_description     #Name1#:Port#Port1# STP Status
    check_command           check_stp!#Ip1#!#Community1#!#Port1#
    check_interval          10
    retry_interval          5
}

define service {
    use                     generic-service
    host_name               localhost
    servicegroups           #ServiceGroup#
    service_description     #Name2#:Port#Port2# STP Status
    check_command           check_stp!#Ip2#!#Community2#!#Port2#
    check_interval          10
    retry_interval          5
}

`
