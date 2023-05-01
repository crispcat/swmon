package main

import (
	"fmt"
	"strconv"
	"strings"
)

type Formatter func(str string) string

func ConstructSwitchTemplate(host *Host) string {

	templates := []string{
		TEMPLATE_NAGIOS_SWITCH,
		TEMPLATE_NAGIOS_PING_SERVICE,
		TEMPLATE_NAGIOS_UPTIME_SERVICE,
	}

	if host.IfsCount > 0 {

		var port uint32
		formatter := func(key string) string {
			switch key {
			case "Port":
				return strconv.FormatUint(uint64(port), 10)
			default:
				return ""
			}
		}

		for _, iface := range host.Interfaces {
			if iface.Operational && iface.SubOid < 256 {
				port = iface.SubOid
				templates = append(templates, FormatTemplate(TEMPLATE_NAGIOS_PORT_LINK_STATUS_SERVICE, formatter))
			}
		}
	}

	return ConstructTemplate(host.DefaultFormatter, templates...)
}

func ConstructConnectionServiceGroup(conn Connection, model *HostsModel) string {

	fr, _ := model.Map.Load(conn.from)
	to, _ := model.Map.Load(conn.to)

	host1 := fr.(*Host)
	host2 := to.(*Host)

	name1 := host1.Ip.String()
	name2 := host2.Ip.String()

	port1 := strconv.FormatUint(uint64(conn.frPort), 10)
	port2 := strconv.FormatUint(uint64(conn.toPort), 10)

	ip1 := host1.Ip.String()
	ip2 := host2.Ip.String()

	community1 := host1.NetworkArgs.SnmpCommunityString
	community2 := host2.NetworkArgs.SnmpCommunityString

	formatter := func(key string) string {
		switch key {

		case "Name1":
			return name1
		case "Port1":
			return port1
		case "Ip1":
			return ip1
		case "Community1":
			return community1

		case "Name2":
			return name2
		case "Port2":
			return port2
		case "Ip2":
			return ip2
		case "Community2":
			return community2

		case "ServiceGroup":
			return fmt.Sprintf(SERVICE_GROUP_FORMAT_STRING, name1, port1, name2, port2)

		default:
			return ""
		}
	}

	return FormatTemplate(TEMPLATE_NAGIOS_CONNECTION_SERVICE_GROUP, formatter)
}

func ConstructTemplate(formatter Formatter, templates ...string) string {

	var sb strings.Builder
	for _, template := range templates {
		formatedTemplate := FormatTemplate(template, formatter)
		sb.WriteString(formatedTemplate)
	}

	return sb.String()
}

func FormatTemplate(template string, formatter Formatter) string {

	if formatter == nil {
		return template
	}

	result := TEMPLATE_REPLACE_REGEX.ReplaceAllStringFunc(template, func(variable string) string {

		injectValue := formatter(strings.Trim(variable, "#"))
		if injectValue == "" {
			// Do nothing
			return variable
		}

		return injectValue
	})

	return result
}
