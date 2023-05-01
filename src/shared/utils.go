package shared

import (
	"fmt"
	"math/big"
	"os"
	"regexp"
	"strconv"
)

var SanitaizerRegex = regexp.MustCompile(`[^a-zA-Z0-9 ]+`)

func Sanitize(str string) string {

	return SanitaizerRegex.ReplaceAllString(str, "")
}

func GetUint32(oid []string, pos int) (uint32, error) {

	var index int
	if pos >= 0 {
		index = pos
	} else {
		index = len(oid) + pos
	}

	num, err := strconv.ParseUint((oid)[index], 10, 32)
	return uint32(num), err
}

func ParseString(value interface{}) string {

	bytes, ok := value.([]byte)
	if !ok {
		return ""
	}

	return string(bytes)
}

func ParseAsUint(value interface{}) string {
	var val int64

	switch value := value.(type) { // shadow
	case int:
		val = int64(value)
	case int8:
		val = int64(value)
	case int16:
		val = int64(value)
	case int32:
		val = int64(value)
	case int64:
		val = value
	case uint:
		val = int64(value)
	case uint8:
		val = int64(value)
	case uint16:
		val = int64(value)
	case uint32:
		val = int64(value)
	case uint64: // beware: int64(MaxUint64) overflow, handle different
		return new(big.Int).SetUint64(value).String()
	case string:
		// for testing and other apps - numbers may appear as strings
		var err error
		if val, err = strconv.ParseInt(value, 10, 64); err != nil {
			return ""
		}
	default:
		return ""
	}

	return big.NewInt(val).String()
}

func PressEnterTo(action string) {

	fmt.Printf("Press ENTER to %s...\n", action)
	var b [1]byte
	os.Stdin.Read(b[:])
}
