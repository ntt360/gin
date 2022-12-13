package conv

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// Struct2Map struct to map，依赖 json tab
func Struct2Map(r interface{}) (s map[string]string, err error) {
	var temp map[string]interface{}
	var result = make(map[string]string)

	bin, err := json.Marshal(r)
	if err != nil {
		return result, err
	}
	if err := json.Unmarshal(bin, &temp); err != nil {
		return nil, err
	}
	for k, v := range temp {
		result[k], err = ToStringE(v)
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

// ToStringE interface to string
func ToStringE(i interface{}) (string, error) {
	switch s := i.(type) {
	case string:
		return s, nil
	case bool:
		return strconv.FormatBool(s), nil
	case float64:
		return strconv.FormatFloat(s, 'f', -1, 64), nil
	case float32:
		return strconv.FormatFloat(float64(s), 'f', -1, 32), nil
	case int:
		return strconv.Itoa(s), nil
	case int64:
		return strconv.FormatInt(s, 10), nil
	case int32:
		return strconv.Itoa(int(s)), nil
	case int16:
		return strconv.FormatInt(int64(s), 10), nil
	case int8:
		return strconv.FormatInt(int64(s), 10), nil
	case uint:
		return strconv.FormatInt(int64(s), 10), nil
	case uint64:
		return strconv.FormatInt(int64(s), 10), nil
	case uint32:
		return strconv.FormatInt(int64(s), 10), nil
	case uint16:
		return strconv.FormatInt(int64(s), 10), nil
	case uint8:
		return strconv.FormatInt(int64(s), 10), nil
	case []byte:
		return string(s), nil
	case nil:
		return "", nil
	case fmt.Stringer:
		return s.String(), nil
	case error:
		return s.Error(), nil
	default:
		return "", fmt.Errorf("unable to cast %#v of type %T to string", i, i)
	}
}

// ToIntE interface to int
func ToIntE(i interface{}) (int64, error) {
	switch s := i.(type) {
	case string:
		v, err := strconv.Atoi(s)
		return int64(v), err
	case int64:
		return s, nil
	case int:
		return int64(s), nil
	case int8:
		return int64(s), nil
	case uint:
		return int64(s), nil
	case uint64:
		return int64(s), nil
	case uint32:
		return int64(s), nil
	case uint16:
		return int64(s), nil
	case uint8:
		return int64(s), nil
	default:
		return 0, fmt.Errorf("unable to cast %#v of type %T to int", i, i)
	}
}

// ToFloatE interface to float64
func ToFloatE(i interface{}) (float64, error) {
	switch s := i.(type) {
	case string:
		return strconv.ParseFloat(s, 64)
	case float64:
		return s, nil
	case float32:
		return float64(s), nil
	case int64:
		return float64(s), nil
	case int:
		return float64(s), nil
	case int32:
		return float64(s), nil
	case int16:
		return float64(s), nil
	case int8:
		return float64(s), nil
	case uint:
		return float64(s), nil
	case uint64:
		return float64(s), nil
	case uint32:
		return float64(s), nil
	case uint16:
		return float64(s), nil
	case uint8:
		return float64(s), nil
	default:
		return 0, fmt.Errorf("unable to cast %#v of type %T to int", i, i)
	}
}
