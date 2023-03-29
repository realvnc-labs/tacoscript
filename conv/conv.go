package conv

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v2"
)

type KeyValues []KeyValue

func (kvs KeyValues) ToEqualSignStrings() []string {
	res := make([]string, 0, len(kvs))
	for _, kv := range kvs {
		res = append(res, kv.ToEqualSignString())
	}

	return res
}

type KeyValue struct {
	Key   string
	Value string
}

var (
	ErrNotANumber           = errors.New("value is not a number")
	ErrFileSizeInvalidUnits = errors.New("file size has invalid units")
	ErrBadBool              = errors.New("failed to parse bool value")
)

func (kv KeyValue) ToEqualSignString() string {
	return fmt.Sprintf("%s=%s", kv.Key, kv.Value)
}

func ConvertToKeyValues(val interface{}, path string) (KeyValues, error) {
	rawKeyValues, ok := val.([]interface{})
	if !ok {
		return []KeyValue{}, fmt.Errorf("key value array expected at '%s' but got '%s'", path, ConvertSourceToJSONStrIfPossible(val))
	}

	res := make([]KeyValue, 0, len(rawKeyValues))

	for _, rawKeyValueI := range rawKeyValues {
		rawKeyValue, ok := rawKeyValueI.(yaml.MapSlice)
		if !ok {
			return []KeyValue{}, fmt.Errorf("wrong key value element at '%s': '%s'", path, ConvertSourceToJSONStrIfPossible(rawKeyValueI))
		}

		for _, item := range rawKeyValue {
			key := item.Key.(string)
			val := item.Value
			res = append(res, KeyValue{
				Key:   key,
				Value: fmt.Sprint(val),
			})
		}
	}

	return res, nil
}

func ConvertToValues(val interface{}) ([]string, error) {
	rawValues, ok := val.([]interface{})
	if !ok {
		return []string{}, fmt.Errorf("values array expected")
	}

	res := make([]string, 0, len(rawValues))

	for _, rawValueI := range rawValues {
		res = append(res, fmt.Sprint(rawValueI))
	}

	return res, nil
}

func ConvertToInt(val interface{}) (num int, err error) {
	valStr := fmt.Sprint(val)
	if !isNumber(valStr) {
		return 0, ErrNotANumber
	}

	num, err = strconv.Atoi(valStr)
	if err != nil {
		return 0, err
	}

	return num, nil
}

func ConvertToBool(val interface{}) (bool, error) {
	boolVal, ok := val.(bool)
	if ok {
		return boolVal, nil
	}

	boolValStr := fmt.Sprint(val)

	switch boolValStr {
	case "":
		return false, nil
	case "false":
		return false, nil
	case "true":
		return true, nil
	case "0":
		return false, nil
	case "1":
		return true, nil
	case "null":
		return false, nil
	default:
		return false, ErrBadBool
	}
}

func ConvertToFileMode(val interface{}) (os.FileMode, error) {
	fileUint, ok := val.(int)
	if ok {
		return os.FileMode(fileUint), nil
	}

	valStr := fmt.Sprint(val)
	i64, err := strconv.ParseInt(valStr, 8, 32)
	if err != nil {
		return 0, fmt.Errorf(`invalid file mode value '%s' at path 'invalid_filemode_path.mode'`, valStr)
	}

	return os.FileMode(i64), nil
}

func ConvertToFileSize(val interface{}) (convertedVal uint64, err error) {
	valStr := fmt.Sprint(val)
	valLen := len(valStr)

	// byte string assumed
	valPart := valStr[:valLen-1]
	units := valStr[valLen-1:]

	if !isNumber(valPart) {
		return 0, ErrNotANumber
	}

	// assume bytes to start with
	multiplier := 1

	// if we have units, work out the multiplier
	if !isNumber(units) {
		multiplier, err = getMultiplierFromUnits(units)
		if err != nil {
			return 0, err
		}
	} else {
		// make sure to include the last digit again
		valPart = valStr
	}

	valNum, err := strconv.ParseUint(valPart, 10, 64)
	if err != nil {
		return 0, err
	}
	convertedVal = valNum * uint64(multiplier)
	return convertedVal, nil
}

func getMultiplierFromUnits(units string) (multiplier int, err error) {
	switch strings.ToLower(units) {
	case "b":
		multiplier = 1
	case "k":
		multiplier = 1024
	case "m":
		multiplier = 1024 * 1024
	case "g":
		multiplier = 1024 * 1024 * 1024
	default:
		return 0, ErrFileSizeInvalidUnits
	}
	return multiplier, nil
}

func isNumber(str string) bool {
	_, err := strconv.Atoi(str)
	return err == nil
}
