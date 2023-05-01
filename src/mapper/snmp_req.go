package main

import (
	goSNMP "github.com/gosnmp/gosnmp"
	"strings"
	utils "swmon_shared"
)

type SnmpResultType uint32

const (
	ResultTypeString SnmpResultType = iota
	ResultTypeInteger
	ResultTypeHardwareAddress
	ResultTypeAuto
)

func (resType SnmpResultType) String() string {
	return []string{"String", "Integer", "HardwareAddress", "Auto"}[resType]
}

type SnmpRequest struct {
	host        *Host
	oid         string
	name        string
	parseToType SnmpResultType
	result      goSNMP.SnmpPDU
}

func CraftSnmpRequest(host *Host, name string, parseToType SnmpResultType) *SnmpRequest {

	return &SnmpRequest{
		host:        host,
		name:        name,
		oid:         MibsGetOid(name),
		parseToType: parseToType,
	}
}

func (req *SnmpRequest) ParseSnmpResult() string {

	if req.result.Value == nil {
		return ""
	}

	switch req.parseToType {

	case ResultTypeAuto:
		return req.assumeParse()

	case ResultTypeString:
		return utils.ParseString(req.result.Value)

	case ResultTypeInteger:
		return utils.ParseAsUint(req.result.Value)

	case ResultTypeHardwareAddress:
		return ParseBytes(req.result.Value, 6, 8, 20)

	default:
		ErrorAll("Non implemented parse result type provided %s")
		return ""
	}
}

func (req *SnmpRequest) assumeParse() string {

	req.parseToType = ResultTypeHardwareAddress
	result := req.parseSnmpResultWithFallbacks(ResultTypeInteger, ResultTypeString)

	return result
}

func (req *SnmpRequest) parseSnmpResultWithFallbacks(fallbacks ...SnmpResultType) string {

	res := req.ParseSnmpResult()
	for i := 0; res == ""; i++ {
		if i >= len(fallbacks) {
			return res
		}
		req.parseToType = fallbacks[i]
		res = req.ParseSnmpResult()
	}

	return res
}

func ParseBytes(value interface{}, assumeLenght ...int) string {

	bytes, ok := value.([]byte)
	if !ok {
		return ""
	}

	ln := len(bytes)
	if ln == 0 {
		return ""
	}

	match := false
	for _, aln := range assumeLenght {
		if ln == aln {
			match = true
			break
		}
	}

	if len(assumeLenght) != 0 && !match {
		return ""
	}

	buf := make([]byte, 0, ln*3-1)
	for i, b := range bytes {
		if i > 0 {
			buf = append(buf, ' ')
		}
		buf = append(buf, hexDigit[b>>4])
		buf = append(buf, hexDigit[b&0xF])
	}

	return string(buf)
}

func (req *SnmpRequest) LogSnmpResult() {

	Verbose("[%s] SNMP GET [%s][%s]. Response:[TYPE:%s][BYTES:%s] --> %s (method=%s) \n",
		req.host.Ip,
		req.name,
		req.oid,
		req.result.Type,
		ParseBytes(req.result.Value),
		req.ParseSnmpResult(),
		req.parseToType)
}

func (req *SnmpRequest) WiteToHost() {

	req.host.Oids[req.oid] = req.ParseSnmpResult()
}

func (req *SnmpRequest) GetSplittedOid() []string {

	splittedOid := strings.Split(req.oid, ".")
	return splittedOid
}

func AssumeNumberOnPos(req *SnmpRequest, oid []string, pos int) (uint32, bool) {

	num, err := utils.GetUint32(oid, pos)
	if err != nil {
		WriteAll("[%s] SNMP GET [%s][%s] OID POS %d IS NOT A NUMBER! MAYBE UNSUPPORTED OID SCHEME",
			req.host.Ip, req.name, req.oid, pos)
		return num, false
	}

	return num, true
}

func AssumePortNumberOnPos(req *SnmpRequest, oid []string, pos int) (uint32, bool) {

	portNumber, ok := AssumeNumberOnPos(req, oid, pos)

	if !ok {
		return portNumber, false
	}

	if portNumber > req.host.IfsCount {
		Verbose("[%s] SNMP GET [%s][%s] OID POS %d IS NOT A VALID PORT NUMBER! MAYBE AN INTERNAL INTERFACE",
			req.host.Ip, req.name, req.oid, pos)
		// resume anyway
	}

	return portNumber, true
}
