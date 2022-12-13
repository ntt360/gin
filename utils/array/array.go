package array

import (
	"errors"
	"reflect"
	"strconv"
	"strings"
)

// Unique for unique element.
func Unique(data interface{}) interface{} {
	typeInfo := reflect.TypeOf(data)
	if typeInfo.Kind() != reflect.Slice {
		panic(errors.New("unique data type must slice"))
	}
	tmp := make(map[interface{}]bool)

	// TODO 有冗余重复代码
	switch typeInfo.String() {
	case "[]uint64":
		var rel []uint64
		originData := data.([]uint64)
		for _, v := range originData {
			if _, ok := tmp[v]; !ok {
				tmp[v] = true
				rel = append(rel, v)
			}
		}
		return rel
	case "[]int64":
		var rel []int64
		originData := data.([]int64)
		for _, v := range originData {
			if _, ok := tmp[v]; !ok {
				tmp[v] = true
				rel = append(rel, v)
			}
		}
		return rel
	case "[]uint":
		var rel []uint
		originData := data.([]uint)
		for _, v := range originData {
			if _, ok := tmp[v]; !ok {
				tmp[v] = true
				rel = append(rel, v)
			}
		}
		return rel
	case "[]int":
		var rel []int
		originData := data.([]int)
		for _, v := range originData {
			if _, ok := tmp[v]; !ok {
				tmp[v] = true
				rel = append(rel, v)
			}
		}
		return rel
	case "[]string":
		var rel []string
		originData := data.([]string)
		for _, v := range originData {
			if _, ok := tmp[v]; !ok {
				tmp[v] = true
				rel = append(rel, v)
			}
		}
		return rel
	}

	return nil
}

// In check string contain any element.
func In(val string, arr []string) bool {
	for _, v := range arr {
		if strings.Contains(v, val) {
			return true
		}
	}

	return false
}

// InEqual check string contain any element.
func InEqual(val string, arr []string) bool {
	for _, v := range arr {
		if v == val {
			return true
		}
	}

	return false
}

// StringSlice convert []int to []string
func StringSlice(inputs []int) []string {
	var rels []string
	for _, input := range inputs {
		rels = append(rels, strconv.Itoa(input))
	}

	return rels
}

// IntSlice convert []string to []int
func IntSlice(inputs []string) []int {
	var rels []int
	for _, input := range inputs {
		mustInt, _ := strconv.Atoi(input)
		rels = append(rels, mustInt)
	}

	return rels
}

// IntEqual if []int equal another []int return true else false
func IntEqual(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}

	for i, v := range a {
		if v != b[i] {
			return false
		}
	}

	return true
}
