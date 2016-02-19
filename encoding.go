package caribou

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
)

// encodeRegister encodes number types into a string so that the value can later be decoded into
// it's original Golang type. If a string is passed in, then it is returned as-is.
func encodeRegister(number interface{}) (encoded string, err error) {
	defer func() {
		if r := recover(); r != nil {
			encoded = ""
			err = errors.New("Could not decode register")
		}
	}()

	switch number.(type) {
	default:
		encoded = number.(string)
	case int, uint, uintptr:
		// Don't allow implementation specific number types.
		err = errors.New("Implementation specific number types are not allowed.")
	case complex64, complex128:
		err = errors.New("Complex numbers aren't supported. Why are you using complex numbers?")
	case int8, int16, int32, int64, uint8, uint16, uint32, uint64:
		encoded = fmt.Sprintf("%s(%d)", reflect.TypeOf(number), number)
	case float32:
		encoded = fmt.Sprintf("float32(%s)", strconv.FormatFloat(float64(number.(float32)), 'E', -1, 32))
	case float64:
		encoded = fmt.Sprintf("float64(%s)", strconv.FormatFloat(number.(float64), 'E', -1, 64))
	}
	return
}

// decodeRegister is the opposite of encodeRegister. It tries to parse out a number type from the
// given string according to our number encoding format. If decoding is not possible, the input is
// returned as-is.
func decodeRegister(str string) (interface{}, error) {
	var r *regexp.Regexp
	var matches []string

	// Decode int8
	r, _ = regexp.Compile("^int8\\((.+)\\)$")
	matches = r.FindStringSubmatch(str)
	if len(matches) > 1 {
		result, err := strconv.ParseInt(matches[1], 10, 8)
		if err != nil {
			return str, err
		}
		return int8(result), nil
	}

	// Decode int16
	r, _ = regexp.Compile("^int16\\((.+)\\)$")
	matches = r.FindStringSubmatch(str)
	if len(matches) > 1 {
		result, err := strconv.ParseInt(matches[1], 10, 16)
		if err != nil {
			return str, err
		}
		return int16(result), nil
	}

	// Decode int32
	r, _ = regexp.Compile("^int32\\((.+)\\)$")
	matches = r.FindStringSubmatch(str)
	if len(matches) > 1 {
		result, err := strconv.ParseInt(matches[1], 10, 32)
		if err != nil {
			return str, err
		}
		return int32(result), nil
	}

	// Decode int64
	r, _ = regexp.Compile("^int64\\((.+)\\)$")
	matches = r.FindStringSubmatch(str)
	if len(matches) > 1 {
		result, err := strconv.ParseInt(matches[1], 10, 64)
		if err != nil {
			return str, err
		}
		return int64(result), nil
	}

	// Decode uint8
	r, _ = regexp.Compile("^uint8\\((.+)\\)$")
	matches = r.FindStringSubmatch(str)
	if len(matches) > 1 {
		result, err := strconv.ParseUint(matches[1], 10, 8)
		if err != nil {
			return str, err
		}
		return uint8(result), nil
	}

	// Decode uint16
	r, _ = regexp.Compile("^uint16\\((.+)\\)$")
	matches = r.FindStringSubmatch(str)
	if len(matches) > 1 {
		result, err := strconv.ParseUint(matches[1], 10, 16)
		if err != nil {
			return str, err
		}
		return uint16(result), nil
	}

	// Decode uint32
	r, _ = regexp.Compile("^uint32\\((.+)\\)$")
	matches = r.FindStringSubmatch(str)
	if len(matches) > 1 {
		result, err := strconv.ParseUint(matches[1], 10, 32)
		if err != nil {
			return str, err
		}
		return uint32(result), nil
	}

	// Decode uint64
	r, _ = regexp.Compile("^uint64\\((.+)\\)$")
	matches = r.FindStringSubmatch(str)
	if len(matches) > 1 {
		result, err := strconv.ParseUint(matches[1], 10, 64)
		if err != nil {
			return str, err
		}
		return uint64(result), nil
	}

	// Decode float32
	r, _ = regexp.Compile("^float32\\((.+)\\)$")
	matches = r.FindStringSubmatch(str)
	if len(matches) > 1 {
		result, err := strconv.ParseFloat(matches[1], 32)
		if err != nil {
			return str, err
		}
		return float32(result), nil
	}

	// Decode float64
	r, _ = regexp.Compile("^float64\\((.+)\\)$")
	matches = r.FindStringSubmatch(str)
	if len(matches) > 1 {
		result, err := strconv.ParseFloat(matches[1], 64)
		if err != nil {
			return str, err
		}
		return float64(result), nil
	}

	return str, nil
}
