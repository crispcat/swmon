// Copyright 2012-2014 The GoSNMP Authors. All rights reserved.  Use of this
// source code is governed by a BSD-style license that can be found in the
// LICENSE file.

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"math/big"
	"net"
	"strconv"
	"strings"
)

// variable struct is used by decodeValue(), which is used for debugging
type variable struct {
	Name  []int
	Type  Asn1BER
	Value interface{}
}

// -- helper functions (mostly) in alphabetical order --------------------------

func decodeValue(data []byte, msg string) (retVal *variable, err error) {
	if LoggingDisabled != true {
		dumpBytes1(data, fmt.Sprintf("decodeValue: %s", msg), 16)
	}
	retVal = new(variable)

	switch Asn1BER(data[0]) {

	case Integer:
		// 0x02. signed
		if LoggingDisabled != true {
			slog.Print("decodeValue: type is Integer")
		}
		length, cursor := parseLength(data)
		var ret int
		var err error
		if ret, err = parseInt(data[cursor:length]); err != nil {
			if LoggingDisabled != true {
				slog.Printf("%v:", err)
			}
			return retVal, fmt.Errorf("bytes: % x err: %v", data, err)
		}
		retVal.Type = Integer
		retVal.Value = ret
	case OctetString:
		// 0x04
		if LoggingDisabled != true {
			slog.Print("decodeValue: type is OctetString")
		}
		length, cursor := parseLength(data)
		retVal.Type = OctetString
		retVal.Value = []byte(data[cursor:length])
	case Null:
		// 0x05
		if LoggingDisabled != true {
			slog.Print("decodeValue: type is Null")
		}
		retVal.Type = Null
		retVal.Value = nil
	case ObjectIdentifier:
		// 0x06
		if LoggingDisabled != true {
			slog.Print("decodeValue: type is ObjectIdentifier")
		}
		rawOid, _, err := parseRawField(data, "OID")
		if err != nil {
			return nil, fmt.Errorf("Error parsing OID Value: %s", err.Error())
		}
		var oid []int
		var ok bool
		if oid, ok = rawOid.([]int); !ok {
			return nil, fmt.Errorf("unable to type assert rawOid |%v| to []int", rawOid)
		}
		retVal.Type = ObjectIdentifier
		retVal.Value = oidToString(oid)
	case IPAddress:
		// 0x40
		if LoggingDisabled != true {
			slog.Print("decodeValue: type is IPAddress")
		}
		retVal.Type = IPAddress
		switch data[1] {
		case 4: // IPv4
			if len(data) < 6 {
				return nil, fmt.Errorf("not enough data for ipv4 address: %x", data)
			}
			retVal.Value = net.IPv4(data[2], data[3], data[4], data[5]).String()
		case 16: // IPv6
			if len(data) < 18 {
				return nil, fmt.Errorf("not enough data for ipv6 address: %x", data)
			}
			d := make(net.IP, 16)
			copy(d, data[2:17])
			retVal.Value = d.String()
		default:
			return nil, fmt.Errorf("got ipaddress len %d, expected 4 or 16", data[1])
		}
	case Counter32:
		// 0x41. unsigned
		if LoggingDisabled != true {
			slog.Print("decodeValue: type is Counter32")
		}
		length, cursor := parseLength(data)
		ret, err := parseUint(data[cursor:length])
		if err != nil {
			if LoggingDisabled != true {
				slog.Printf("decodeValue: err is %v", err)
			}
			break
		}
		retVal.Type = Counter32
		retVal.Value = ret
	case Gauge32:
		// 0x42. unsigned
		if LoggingDisabled != true {
			slog.Print("decodeValue: type is Gauge32")
		}
		length, cursor := parseLength(data)
		ret, err := parseUint(data[cursor:length])
		if err != nil {
			if LoggingDisabled != true {
				slog.Printf("decodeValue: err is %v", err)
			}
			break
		}
		retVal.Type = Gauge32
		retVal.Value = ret
	case TimeTicks:
		// 0x43
		if LoggingDisabled != true {
			slog.Print("decodeValue: type is TimeTicks")
		}
		length, cursor := parseLength(data)
		ret, err := parseInt(data[cursor:length])
		if err != nil {
			if LoggingDisabled != true {
				slog.Printf("decodeValue: err is %v", err)
			}
			break
		}
		retVal.Type = TimeTicks
		retVal.Value = ret
	case Counter64:
		// 0x46
		if LoggingDisabled != true {
			slog.Print("decodeValue: type is Counter64")
		}
		length, cursor := parseLength(data)
		ret, err := parseInt64(data[cursor:length])
		if err != nil {
			if LoggingDisabled != true {
				slog.Printf("decodeValue: err is %v", err)
			}
			break
		}
		retVal.Type = Counter64
		retVal.Value = ret
	case NoSuchObject:
		// 0x80
		if LoggingDisabled != true {
			slog.Print("decodeValue: type is NoSuchObject")
		}
		retVal.Type = NoSuchObject
		retVal.Value = nil
	case NoSuchInstance:
		// 0x81
		if LoggingDisabled != true {
			slog.Print("decodeValue: type is NoSuchInstance")
		}
		retVal.Type = NoSuchInstance
		retVal.Value = nil
	case EndOfMibView:
		// 0x82
		if LoggingDisabled != true {
			slog.Print("decodeValue: type is EndOfMibView")
		}
		retVal.Type = EndOfMibView
		retVal.Value = nil
	default:
		if LoggingDisabled != true {
			slog.Print("decodeValue: type %x isn't implemented", data[0])
		}
		retVal.Type = UnknownType
		retVal.Value = nil
	}
	if LoggingDisabled != true {
		slog.Printf("decodeValue: value is %#v", retVal.Value)
	}
	return
}

// dump bytes in a format similar to Wireshark
func dumpBytes1(data []byte, msg string, maxlength int) {
	var buffer bytes.Buffer
	buffer.WriteString(msg)
	length := maxlength
	if len(data) < maxlength {
		length = len(data)
	}
	length *= 2 //One Byte Symobls Two Hex
	hexStr := hex.EncodeToString(data)
	for i := 0; length >= i+16; i += 16 {
		buffer.WriteString("\n")
		buffer.WriteString(strconv.Itoa(i / 2))
		buffer.WriteString("\t")
		buffer.WriteString(hexStr[i : i+2])
		buffer.WriteString(" ")
		buffer.WriteString(hexStr[i+2 : i+4])
		buffer.WriteString(" ")
		buffer.WriteString(hexStr[i+4 : i+6])
		buffer.WriteString(" ")
		buffer.WriteString(hexStr[i+6 : i+8])
		buffer.WriteString(" ")
		buffer.WriteString(hexStr[i+8 : i+10])
		buffer.WriteString(" ")
		buffer.WriteString(hexStr[i+10 : i+12])
		buffer.WriteString(" ")
		buffer.WriteString(hexStr[i+12 : i+14])
		buffer.WriteString(" ")
		buffer.WriteString(hexStr[i+14 : i+16])
	}
	leftOver := length % 16
	if leftOver != 0 {
		buffer.WriteString("\n")
		buffer.WriteString(strconv.Itoa((length - leftOver) / 2))
		buffer.WriteString("\t")
		for i := 0; leftOver >= i+2; i += 2 {
			buffer.WriteString(hexStr[i : i+2])
			buffer.WriteString(" ")
		}
	}
	buffer.WriteString("\n")
	slog.Print(buffer.String())
}

// dump bytes in one row, up to about screen width. Returns a string
// rather than (dumpBytes1) writing to debugging log.
func dumpBytes2(desc string, bb []byte, cursor int) string {
	cursor = cursor - 4 // give some context to dump
	if cursor < 0 {
		cursor = 0
	}
	result := desc
	for i, b := range bb[cursor:] {
		if i > 30 { // about screen width...
			break
		}
		result += fmt.Sprintf(" %02x", b)
	}
	return result
}

func marshalBase128Int(out *bytes.Buffer, n int64) (err error) {
	if n == 0 {
		err = out.WriteByte(0)
		return
	}

	l := 0
	for i := n; i > 0; i >>= 7 {
		l++
	}

	for i := l - 1; i >= 0; i-- {
		o := byte(n >> uint(i*7))
		o &= 0x7f
		if i != 0 {
			o |= 0x80
		}
		err = out.WriteByte(o)
		if err != nil {
			return
		}
	}

	return nil
}

// marshalInt16 builds a byte representation of
// a 16 bit int in BigEndian form.
// TODO add an error return, like other marshals
func marshalInt16(value int) (rs []byte) {
	if value <= 0xff {
		rs = []byte{byte(value)}
		return
	}
	if value > 0xff && value <= 0xffff {
		rs = []byte{byte(((value >> 8) & 0xff)), byte((value & 0xff))}
		return
	}
	return
}

// marshalLength builds a byte representation of length
//
// http://luca.ntop.org/Teaching/Appunti/asn1.html
//
// Length octets. There are two forms: short (for lengths between 0 and 127),
// and long definite (for lengths between 0 and 2^1008 -1).
//
//   - Short form. One octet. Bit 8 has value "0" and bits 7-1 give the length.
//   - Long form. Two to 127 octets. Bit 8 of first octet has value "1" and bits
//     7-1 give the number of additional length octets. Second and following
//     octets give the length, base 256, most significant digit first.
func marshalLength(length int) ([]byte, error) {
	// more convenient to pass length as int than uint64. Therefore check < 0
	if length < 0 {
		return nil, fmt.Errorf("length must be greater than zero")
	} else if length < 127 {
		return []byte{byte(length)}, nil
	}

	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, uint64(length))
	if err != nil {
		return nil, err
	}
	bufBytes := buf.Bytes()

	// strip leading zeros
	for idx, octect := range bufBytes {
		if octect != 00 {
			bufBytes = bufBytes[idx:len(bufBytes)]
			break
		}
	}

	header := []byte{byte(128 | len(bufBytes))}
	return append(header, bufBytes...), nil
}

func marshalObjectIdentifier(oid []int) (ret []byte, err error) {
	out := new(bytes.Buffer)
	if len(oid) < 2 || oid[0] > 6 || oid[1] >= 40 {
		return nil, errors.New("invalid object identifier")
	}

	err = out.WriteByte(byte(oid[0]*40 + oid[1]))
	if err != nil {
		return
	}
	for i := 2; i < len(oid); i++ {
		err = marshalBase128Int(out, int64(oid[i]))
		if err != nil {
			return
		}
	}

	ret = out.Bytes()
	return
}

func marshalOID(oid string) ([]byte, error) {
	var err error

	// Encode the oid
	oid = strings.Trim(oid, ".")
	oidParts := strings.Split(oid, ".")
	oidBytes := make([]int, len(oidParts))

	// Convert the string OID to an array of integers
	for i := 0; i < len(oidParts); i++ {
		oidBytes[i], err = strconv.Atoi(oidParts[i])
		if err != nil {
			return nil, fmt.Errorf("Unable to parse OID: %s\n", err.Error())
		}
	}

	mOid, err := marshalObjectIdentifier(oidBytes)

	if err != nil {
		return nil, fmt.Errorf("Unable to marshal OID: %s\n", err.Error())
	}

	return mOid, err
}

func oidToString(oid []int) (ret string) {
	oidAsString := make([]string, len(oid)+1)

	// used for appending of the first dot
	oidAsString[0] = ""
	for i := range oid {
		oidAsString[i+1] = strconv.Itoa(oid[i])
	}

	return strings.Join(oidAsString, ".")
}

// parseBase128Int parses a base-128 encoded int from the given offset in the
// given byte slice. It returns the value and the new offset.
func parseBase128Int(bytes []byte, initOffset int) (ret, offset int, err error) {
	offset = initOffset
	for shifted := 0; offset < len(bytes); shifted++ {
		if shifted > 4 {
			err = fmt.Errorf("Structural Error: base 128 integer too large")
			return
		}
		ret <<= 7
		b := bytes[offset]
		ret |= int(b & 0x7f)
		offset++
		if b&0x80 == 0 {
			return
		}
	}
	err = fmt.Errorf("Syntax Error: truncated base 128 integer")
	return
}

// parseBitString parses an ASN.1 bit string from the given byte slice and returns it.
func parseBitString(bytes []byte) (ret BitStringValue, err error) {
	if len(bytes) == 0 {
		err = errors.New("zero length BIT STRING")
		return
	}
	paddingBits := int(bytes[0])
	if paddingBits > 7 ||
		len(bytes) == 1 && paddingBits > 0 ||
		bytes[len(bytes)-1]&((1<<bytes[0])-1) != 0 {
		err = errors.New("invalid padding bits in BIT STRING")
		return
	}
	ret.BitLength = (len(bytes)-1)*8 - paddingBits
	ret.Bytes = bytes[1:]
	return
}

// parseInt64 treats the given bytes as a big-endian, signed integer and
// returns the result.
func parseInt64(bytes []byte) (ret int64, err error) {
	if len(bytes) > 8 {
		// We'll overflow an int64 in this case.
		err = errors.New("integer too large")
		return
	}
	for bytesRead := 0; bytesRead < len(bytes); bytesRead++ {
		ret <<= 8
		ret |= int64(bytes[bytesRead])
	}

	// Shift up and down in order to sign extend the result.
	ret <<= 64 - uint8(len(bytes))*8
	ret >>= 64 - uint8(len(bytes))*8
	return
}

// parseInt treats the given bytes as a big-endian, signed integer and returns
// the result.
func parseInt(bytes []byte) (int, error) {
	ret64, err := parseInt64(bytes)
	if err != nil {
		return 0, err
	}
	if ret64 != int64(int(ret64)) {
		return 0, errors.New("integer too large")
	}
	return int(ret64), nil
}

// parseLength parses and calculates an snmp packet length
//
// http://luca.ntop.org/Teaching/Appunti/asn1.html
//
// Length octets. There are two forms: short (for lengths between 0 and 127),
// and long definite (for lengths between 0 and 2^1008 -1).
//
//   - Short form. One octet. Bit 8 has value "0" and bits 7-1 give the length.
//   - Long form. Two to 127 octets. Bit 8 of first octet has value "1" and bits
//     7-1 give the number of additional length octets. Second and following
//     octets give the length, base 256, most significant digit first.
func parseLength(bytes []byte) (length int, cursor int) {
	if len(bytes) <= 2 {
		// handle null octet strings ie "0x04 0x00"
		cursor = 1
		length = 2
	} else if int(bytes[1]) <= 127 {
		length = int(bytes[1])
		length += 2
		cursor += 2
	} else {
		numOctets := int(bytes[1]) & 127
		for i := 0; i < numOctets; i++ {
			length <<= 8
			length += int(bytes[2+i])
		}
		length += 2 + numOctets
		cursor += 2 + numOctets
	}
	return length, cursor
}

// parseObjectIdentifier parses an OBJECT IDENTIFIER from the given bytes and
// returns it. An object identifier is a sequence of variable length integers
// that are assigned in a hierarchy.
func parseObjectIdentifier(bytes []byte) (s []int, err error) {
	if len(bytes) == 0 {
		err = fmt.Errorf("zero length OBJECT IDENTIFIER")
		return
	}

	// In the worst case, we get two elements from the first byte (which is
	// encoded differently) and then every varint is a single byte long.
	s = make([]int, len(bytes)+1)

	// The first byte is 40*value1 + value2:
	s[0] = int(bytes[0]) / 40
	s[1] = int(bytes[0]) % 40
	i := 2
	for offset := 1; offset < len(bytes); i++ {
		var v int
		v, offset, err = parseBase128Int(bytes, offset)
		if err != nil {
			return
		}
		s[i] = v
	}
	s = s[0:i]
	return
}

func parseRawField(data []byte, msg string) (interface{}, int, error) {
	dumpBytes1(data, fmt.Sprintf("parseRawField: %s", msg), 16)

	switch Asn1BER(data[0]) {
	case Integer:
		length, cursor := parseLength(data)
		i, err := parseInt(data[cursor:length])
		if err != nil {
			return nil, 0, fmt.Errorf("Unable to parse raw INTEGER: %x err: %v", data, err)
		}
		return i, length, nil
	case OctetString:
		length, cursor := parseLength(data)
		return string(data[cursor:length]), length, nil
	case ObjectIdentifier:
		length := int(data[1])
		oid, err := parseObjectIdentifier(data[2 : 2+length])
		return oid, length + 2, err
	default:
		return nil, 0, fmt.Errorf("Unknown field type: %x\n", data[0])
	}

	return nil, 0, nil
}

func parseUint16(content []byte) int {
	number := uint8(content[1]) | uint8(content[0])<<8
	return int(number)
}

// parseUint64 treats the given bytes as a big-endian, unsigned integer and returns
// the result.
func parseUint64(bytes []byte) (ret uint64, err error) {
	if len(bytes) > 8 {
		// We'll overflow a uint64 in this case.
		err = errors.New("integer too large")
		return
	}
	for bytesRead := 0; bytesRead < len(bytes); bytesRead++ {
		ret <<= 8
		ret |= uint64(bytes[bytesRead])
	}
	return
}

// parseUint treats the given bytes as a big-endian, signed integer and returns
// the result.
func parseUint(bytes []byte) (uint, error) {
	ret64, err := parseUint64(bytes)
	if err != nil {
		return 0, err
	}
	if ret64 != uint64(uint(ret64)) {
		return 0, errors.New("integer too large")
	}
	return uint(ret64), nil
}

// Issue 4389: math/big: add SetUint64 and Uint64 functions to *Int
//
// uint64ToBigInt copied from: http://github.com/cznic/mathutil/blob/master/mathutil.go#L341
//
// replace with Uint64ToBigInt or equivalent when using Go 1.1

var uint64ToBigIntDelta big.Int

func init() {
	uint64ToBigIntDelta.SetBit(&uint64ToBigIntDelta, 63, 1)
}

func uint64ToBigInt(n uint64) *big.Int {
	if n <= math.MaxInt64 {
		return big.NewInt(int64(n))
	}

	y := big.NewInt(int64(n - uint64(math.MaxInt64) - 1))
	return y.Add(y, &uint64ToBigIntDelta)
}

// -- Bit String ---------------------------------------------------------------

// BitStringValue is the structure to use when you want an ASN.1 BIT STRING type. A
// bit string is padded up to the nearest byte in memory and the number of
// valid bits is recorded. Padding bits will be zero.
type BitStringValue struct {
	Bytes     []byte // bits packed into bytes.
	BitLength int    // length in bits.
}

// At returns the bit at the given index. If the index is out of range it
// returns false.
func (b BitStringValue) At(i int) int {
	if i < 0 || i >= b.BitLength {
		return 0
	}
	x := i / 8
	y := 7 - uint(i%8)
	return int(b.Bytes[x]>>y) & 1
}

// RightAlign returns a slice where the padding bits are at the beginning. The
// slice may share memory with the BitString.
func (b BitStringValue) RightAlign() []byte {
	shift := uint(8 - (b.BitLength % 8))
	if shift == 8 || len(b.Bytes) == 0 {
		return b.Bytes
	}

	a := make([]byte, len(b.Bytes))
	a[0] = b.Bytes[0] >> shift
	for i := 1; i < len(b.Bytes); i++ {
		a[i] = b.Bytes[i-1] << (8 - shift)
		a[i] |= b.Bytes[i] >> shift
	}

	return a
}

// -- SnmpVersion --------------------------------------------------------------

func (s SnmpVersion) String() string {
	if s == Version1 {
		return "1"
	}
	return "2c"
}
