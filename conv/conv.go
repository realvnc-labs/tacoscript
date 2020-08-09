package conv

import "fmt"

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
		rawKeyValue, ok := rawKeyValueI.(map[string]interface{})
		if !ok {
			return []KeyValue{}, fmt.Errorf("wrong key value element at '%s': '%s'", path, ConvertSourceToJSONStrIfPossible(rawKeyValueI))
		}

		for key, val := range rawKeyValue {
			res = append(res, KeyValue{
				Key:   key,
				Value: fmt.Sprint(val),
			})
		}
	}

	return res, nil
}

func ConvertToValues(val interface{}, path string) ([]string, error) {
	rawValues, ok := val.([]interface{})
	if !ok {
		return []string{}, fmt.Errorf("values array expected at '%s' but got '%s'", path, ConvertSourceToJSONStrIfPossible(val))
	}

	res := make([]string, 0, len(rawValues))

	for _, rawValueI := range rawValues {
		rawValue, ok := rawValueI.([]interface{})
		if !ok {
			return []string{}, fmt.Errorf("values array expected at '%s' but got '%s'", path, ConvertSourceToJSONStrIfPossible(rawValueI))
		}

		for _, val := range rawValue {
			res = append(res, fmt.Sprint(val))
		}
	}

	return res, nil
}
