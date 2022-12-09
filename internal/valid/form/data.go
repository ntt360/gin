package form

import (
	"reflect"
	"strconv"
	"strings"
)

type Data map[string]any

func (d Data) Value(keyIdx string) (reflect.Value, bool) {
	if len(keyIdx) == 0 {
		return reflect.Value{}, false
	}
	val := reflect.ValueOf(d)

BEGIN:
	if keyIdx == "" {
		return val, true
	}

	t := val.Type().Kind()
	switch t {
	case reflect.Map:
		var k, l string
		idx := strings.Index(keyIdx, namespaceSeparator)
		if idx >= 0 {
			k = keyIdx[0:idx]
			l = keyIdx[idx+1:]
		} else {
			k = keyIdx
		}

		for _, item := range val.MapKeys() {
			if item.Interface().(string) == k {
				val = val.MapIndex(item)
				keyIdx = l
				goto BEGIN
			}
		}

		// if the key not found
		return val, false
	case reflect.Interface, reflect.Ptr:
		val = val.Elem()
		goto BEGIN
	case reflect.Array, reflect.Slice:
		var pos = keyIdx
		var l string

		idx := strings.Index(keyIdx, namespaceSeparator)
		if idx >= 0 {
			pos = keyIdx[:idx]
			l = keyIdx[idx+1:]
		}

		posIndex, _ := strconv.Atoi(pos)
		if posIndex >= val.Len() {
			return val, false
		}

		val = val.Index(posIndex)
		keyIdx = l
		goto BEGIN

	default:
		return val, false
	}
}
