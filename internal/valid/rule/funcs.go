package rule

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/leodido/go-urn"
	"math"
	"net"
	"net/url"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

var (
	oneofValsCache       = map[string][]string{}
	oneofValsCacheRWLock = sync.RWMutex{}
)

// isRegexMatch regex match, this validator must be used with another CurTag: pattern
//
// Demo:
//
//	type Params struct {
//			Name string `json:"name" pattern:"\\w{1,3}" binding:"regex"`
//
//			// array elem regex match
//			List []string `json:"list" pattern:"^\\w{1,3}$" binding:"dive,regex"`
//	}
func isRegexMatch(fl Field) bool {
	if !fl.Exist() {
		return false
	}

	pattern := fl.Tag().Get(KeyRegex)
	if pattern == "" {
		pattern = fl.InheritTag().Get(KeyRegex)
		if pattern == "" {
			panic("the regex validator must be used with pattern CurTag")
		}
	}

	reg, err := regexp.Compile(pattern)
	if err != nil {
		panic(fmt.Sprintf("the regex pattern not invalid: %s", pattern))
	}

	return reg.MatchString(fl.ToString())
}

// isMobile valid china all isp mobile number, can  begin with +86 or not
func isMobile(fl Field) bool {
	if !fl.Exist() {
		return false
	}

	rel, ok := fl.Str()
	if !ok {
		return false
	}

	return mobileRegex.MatchString(rel)
}

// isIdCard valid the string whether is china id number
func isIdCard(fl Field) bool {
	if !fl.Exist() {
		return false
	}

	rel, ok := fl.Str()
	if !ok {
		return false
	}

	idCard := strings.ToUpper(rel)
	var reg, err = regexp.Compile(`^[0-9]{17}[0-9X]$`)
	if err != nil {
		return false
	}

	if !reg.Match([]byte(idCard)) {
		return false
	}

	var sum, num int
	for index, c := range idCard {
		if index != 17 {
			if v, err := strconv.Atoi(string(c)); err == nil {
				//计算加权因子
				sum += int(math.Pow(2, float64(17-index))) % 11 * v
			} else {
				return false
			}
		}
	}

	num = (12 - (sum % 11)) % 11

	checkCode := idCard[len(idCard)-1:]
	if num < 10 {
		cc, _ := strconv.Atoi(checkCode)
		return num == cc
	}
	return checkCode == "X"
}

// hasValue required rule check
func hasValue(fl Field) bool {
	return fl.Exist()
}

// hasMinOf is the validation function for validating if the current field's value is greater than or equal to the param's value.
func hasMinOf(fl Field) bool {
	return isGte(fl)
}

func getRuleIntVal(fl Field) int {
	p, err := strconv.Atoi(fl.RuleVal())
	if err != nil {
		panic(fmt.Errorf("the Key %s valid rule %s value %s parse err: %s", fl.Key(), fl.RuleName(), fl.RuleVal(), err.Error()))
	}

	return p
}

// isGte is the validation function for validating if the current field's value is greater than or equal to the param's value.
func isGte(fl Field) bool {
	if !fl.Exist() { // 数据字段不存在直接返回
		return false
	}

	switch fl.Self().Kind() { // 定义的 Struct 想要的数据类型
	case reflect.String:
		p := getRuleIntVal(fl)

		v, ok := fl.Str()
		if ok {
			return int64(utf8.RuneCountInString(v)) >= int64(p)
		}

		return false

	case reflect.Slice, reflect.Map, reflect.Array:
		p := getRuleIntVal(fl)

		if fl.IsArray() || fl.IsObject() {
			return int64(fl.Len()) >= int64(p)
		}

		return false

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		p := getRuleIntVal(fl)

		v, ok := fl.Int()
		if !ok {
			return false
		}

		return v >= int64(p)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		p := getRuleIntVal(fl)

		v, ok := fl.Uint()
		if !ok {
			return false
		}

		return v >= uint64(p)

	case reflect.Float32, reflect.Float64:
		p := getRuleIntVal(fl)

		v, ok := fl.Float()
		if !ok {
			return false
		}

		return v >= float64(p)

	case reflect.Struct:
		if fl.Self().ConvertibleTo(TimeType) {
			now := time.Now().UTC()
			v, ok := fl.Str()
			if !ok {
				return false
			}

			t, e := FieldTimeVal(v, fl.Tag())
			if e != nil {
				return false
			}

			// 此时，p 值没参与，没有任何用处
			return t.After(now) || t.Equal(now)
		}
	}

	panic(fmt.Sprintf("field type %s not support validator: %s", fl.Self().Name(), fl.RuleName()))
}

func FieldTimeVal(val string, tag reflect.StructTag) (time.Time, error) {
	timeFormat := tag.Get("time_format")
	if timeFormat == "" {
		timeFormat = time.RFC3339
	}

	switch tf := strings.ToLower(timeFormat); tf {
	case "unix", "unixmilli", "unixmicro", "unixnano":
		tv, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return time.Time{}, err
		}

		t := time.Unix(0, tv)
		switch tf {
		case "unix":
			t = time.Unix(tv, 0)
		case "unixmilli":
			t = time.Unix(tv/1e3, (tv%1e3)*1e6)
		case "unixmicro":
			t = time.Unix(tv/1e6, (tv%1e6)*1e3)
		}

		return t, nil
	}

	if val == "" {
		return time.Time{}, errors.New("val is empty")
	}

	l := time.Local
	if isUTC, _ := strconv.ParseBool(tag.Get("time_utc")); isUTC {
		l = time.UTC
	}

	if locTag := tag.Get("time_location"); locTag != "" {
		loc, err := time.LoadLocation(locTag)
		if err != nil {
			return time.Time{}, err
		}
		l = loc
	}

	return time.ParseInLocation(timeFormat, val, l)
}

// requiredIf is the validation function
// The field under validation must be present and not empty only if all the other specified fields are equal to the value following with the specified field.
func requiredIf(fl Field) bool {
	params := parseOneOfParam2(fl.RuleVal())

	// 必须3个一组出现
	if len(params)%3 != 0 {
		panic(fmt.Sprintf("Bad param number for required_if %s", fl.Key()))
	}

	for i := 0; i < len(params); i += 3 {
		fieldName := params[i]   // 字段
		fieldCond := params[i+1] // 比较条件
		filedVal := params[i+2]  // 值

		// 一次判断当前fl 和相对的 filed 值是否匹配
		if !requireCheckFieldValue(fl, fieldName, fieldCond, filedVal) {
			return false
		}
	}

	return hasValue(fl)
}

func parseOneOfParam2(s string) []string {
	oneofValsCacheRWLock.RLock()
	vals, ok := oneofValsCache[s]
	oneofValsCacheRWLock.RUnlock()
	if !ok {
		oneofValsCacheRWLock.Lock()
		vals = splitParamsRegex.FindAllString(s, -1)
		for i := 0; i < len(vals); i++ {
			vals[i] = strings.Replace(vals[i], "'", "", -1)
		}
		oneofValsCache[s] = vals
		oneofValsCacheRWLock.Unlock()
	}
	return vals
}

type NumericT interface {
	float64 | int64 | int
}

func requireCheckFieldValue(fl Field, fieldName string, fieldCond string, filedVal string) bool {
	findField, found := fl.StructField(fl.Parent(), fieldName)
	if !found {
		return false
	}

	if !findField.Exist() {
		return false
	}

	switch findField.Self().Kind() {

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64:
		val, _ := strconv.Atoi(filedVal)

		if !findField.Exist() {
			return false
		}

		v, ok := findField.Int()
		if !ok {
			return false
		}

		return condCompare(int64(val), fieldCond, v)

	case reflect.Slice, reflect.Map, reflect.Array:
		if !(fl.IsArray() || fl.IsObject()) {
			return false
		}

		val, _ := strconv.Atoi(filedVal)
		return condCompare(val, fieldCond, fl.Len())

	case reflect.Bool:
		b, _ := strconv.ParseBool(filedVal)
		if fieldCond != "eq" {
			return false
		}

		fb, ok := findField.Bool()
		if !ok {
			return false
		}

		return b == fb

	default: // default string
		str, ok := findField.Str()
		if !ok {
			return false
		}

		if fieldCond != "eq" {
			return false
		}

		return filedVal == str
	}
}

func condCompare[T NumericT](val T, fieldCond string, match T) bool {
	switch fieldCond {
	case "gt":
		return val > match
	case "gte":
		return val >= match
	case "eq":
		return val == match
	case "lt":
		return val < match
	case "ne":
		return val != match
	default:
		return val <= match
	}
}

func isURLEncoded(fl Field) bool {
	str, ok := fl.Str()
	if !ok {
		return false
	}

	return uRLEncodedRegex.MatchString(str)
}

func isHTMLEncoded(fl Field) bool {
	str, ok := fl.Str()
	if !ok {
		return false
	}

	return hTMLEncodedRegex.MatchString(str)
}

func isHTML(fl Field) bool {
	str, ok := fl.Str()
	if !ok {
		return false
	}

	return hTMLRegex.MatchString(str)
}

func isOneOf(fl Field) bool {
	vals := parseOneOfParam2(fl.RuleVal())

	var v string
	switch fl.Self().Kind() {
	case reflect.String:
		str, ok := fl.Str()
		if !ok {
			return false
		}

		v = str

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		rel, ok := fl.Int()
		if !ok {
			return false
		}

		v = strconv.FormatInt(rel, 10)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		rel, ok := fl.Uint()
		if !ok {
			return false
		}

		v = strconv.FormatUint(rel, 10)
	default:
		panic(fmt.Sprintf("Bad field type %s", fl.Raw()))
	}

	for i := 0; i < len(vals); i++ {
		if vals[i] == v {
			return true
		}
	}
	return false
}

// isUnique is the validation function for validating if each array|slice|map value is unique
func isUnique(fl Field) bool {
	param := fl.RuleVal()
	v := reflect.ValueOf(struct{}{})

	oldIdx := fl.CurDataIndex()
	var oldIdsArr []string
	if len(oldIdx) > 0 {
		oldIdsArr = strings.Split(oldIdx, ".")
	}

	switch fl.Self().Kind() {
	case reflect.Slice, reflect.Array:
		elem := fl.Self().Elem()
		if elem.Kind() == reflect.Ptr {
			elem = elem.Elem()
		}

		var t reflect.Type
		var sf reflect.StructField

		if param == "" {
			t = elem
		} else { // check the value of struct key weather unique
			var ok bool
			sf, ok = elem.FieldByName(fl.RuleVal())
			if !ok {
				panic(fmt.Sprintf("Bad field name %s", param))
			}
			sfTyp := sf.Type
			if sfTyp.Kind() == reflect.Ptr {
				sfTyp = sfTyp.Elem()
			}

			t = sfTyp
		}

		// check the array elements value type
		switch t.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			kTyp := reflect.TypeOf(int64(0))
			m := reflect.MakeMap(reflect.MapOf(kTyp, v.Type()))

			for i := 0; i < fl.Len(); i++ {
				newIds := append(oldIdsArr, strconv.Itoa(i))
				if len(fl.RuleVal()) != 0 {
					newIds = append(newIds, fl.RuleVal())
				}

				fl.SetDataIdx(strings.Join(newIds, "."))

				rel, ok := fl.Int()
				if !ok {
					return false
				}
				m.SetMapIndex(reflect.ValueOf(rel), v)

				// reset data idx
				fl.SetDataIdx(oldIdx)
			}

			return fl.Len() == m.Len()
		case reflect.Float32, reflect.Float64:
			kTyp := reflect.TypeOf(float64(0))
			m := reflect.MakeMap(reflect.MapOf(kTyp, v.Type()))

			for i := 0; i < fl.Len(); i++ {
				newIds := append(oldIdsArr, strconv.Itoa(i))
				if len(fl.RuleVal()) != 0 {
					newIds = append(newIds, fl.RuleVal())
				}

				fl.SetDataIdx(strings.Join(newIds, "."))

				rel, ok := fl.Float()
				if !ok {
					return false
				}
				m.SetMapIndex(reflect.ValueOf(rel), v)

				// reset data idx
				fl.SetDataIdx(oldIdx)
			}

			return fl.Len() == m.Len()
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			kTyp := reflect.TypeOf(uint64(0))
			m := reflect.MakeMap(reflect.MapOf(kTyp, v.Type()))

			for i := 0; i < fl.Len(); i++ {
				newIds := append(oldIdsArr, strconv.Itoa(i))
				if len(fl.RuleVal()) != 0 {
					newIds = append(newIds, fl.RuleVal())
				}
				fl.SetDataIdx(strings.Join(newIds, "."))

				rel, ok := fl.Uint()
				if !ok {
					return false
				}
				m.SetMapIndex(reflect.ValueOf(rel), v)

				// reset data idx
				fl.SetDataIdx(oldIdx)
			}

			return fl.Len() == m.Len()
		case reflect.Bool:
			kTyp := reflect.TypeOf(false)
			m := reflect.MakeMap(reflect.MapOf(kTyp, v.Type()))

			for i := 0; i < fl.Len(); i++ {
				newIds := append(oldIdsArr, strconv.Itoa(i))
				if len(fl.RuleVal()) != 0 {
					newIds = append(newIds, fl.RuleVal())
				}
				fl.SetDataIdx(strings.Join(newIds, "."))

				rel, ok := fl.Bool()
				if !ok {
					return false
				}

				m.SetMapIndex(reflect.ValueOf(rel), v)

				// reset data idx
				fl.SetDataIdx(oldIdx)
			}

			return fl.Len() == m.Len()

		case reflect.Interface, reflect.String:
			kTyp := reflect.TypeOf("")
			m := reflect.MakeMap(reflect.MapOf(kTyp, v.Type()))

			for i := 0; i < fl.Len(); i++ {
				// tmp rebuild the data index for find data value
				newDataIdx := append(oldIdsArr, strconv.Itoa(i))
				if len(fl.RuleVal()) != 0 {
					pVal := Param(sf.Tag, fl.ParamKey())
					if pVal == "" || pVal == "-" {
						panic(fmt.Sprintf("the strcut key must define %s or not use - ignore", sf.Name))
					}

					newDataIdx = append(newDataIdx, pVal)
				}
				fl.SetDataIdx(strings.Join(newDataIdx, "."))

				// try to find data by new data index position
				m.SetMapIndex(reflect.ValueOf(fl.Raw()), v)

				// reset data idx after used
				fl.SetDataIdx(oldIdx)
			}
			return fl.Len() == m.Len()
		default:
			panic(fmt.Sprintf("Bad field type %s", elem.Kind().String()))
		}

	case reflect.Map: // map value type
		mTyp := fl.Self().Elem()
		if mTyp.Kind() == reflect.Ptr {
			mTyp = mTyp.Elem()
		}

		var m reflect.Value
		for _, k := range fl.MapKeys() {
			switch mTyp.Kind() {
			case reflect.String, reflect.Interface, reflect.Struct, reflect.Array, reflect.Slice, reflect.Map:
				m = reflect.MakeMap(reflect.MapOf(mTyp, v.Type()))
				newIds := append(oldIdsArr, k)
				fl.SetDataIdx(strings.Join(newIds, "."))
				m.SetMapIndex(reflect.ValueOf(fl.Raw()), v)
				fl.SetDataIdx(oldIdx)
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				m = reflect.MakeMap(reflect.MapOf(reflect.TypeOf(int64(0)), v.Type()))
				newIds := append(oldIdsArr, k)
				fl.SetDataIdx(strings.Join(newIds, "."))

				rel, ok := fl.Int()
				if !ok {
					return false
				}

				m.SetMapIndex(reflect.ValueOf(rel), v)
				fl.SetDataIdx(oldIdx)
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
				reflect.Uintptr:

				m = reflect.MakeMap(reflect.MapOf(reflect.TypeOf(uint64(0)), v.Type()))
				newIds := append(oldIdsArr, k)
				fl.SetDataIdx(strings.Join(newIds, "."))

				rel, ok := fl.Uint()
				if !ok {
					return false
				}

				m.SetMapIndex(reflect.ValueOf(rel), v)
				fl.SetDataIdx(oldIdx)
			case reflect.Float32, reflect.Float64:
				m = reflect.MakeMap(reflect.MapOf(reflect.TypeOf(float64(0)), v.Type()))
				newIds := append(oldIdsArr, k)
				fl.SetDataIdx(strings.Join(newIds, "."))

				rel, ok := fl.Float()
				if !ok {
					return false
				}

				m.SetMapIndex(reflect.ValueOf(rel), v)
				fl.SetDataIdx(oldIdx)
			case reflect.Bool:
				m = reflect.MakeMap(reflect.MapOf(reflect.TypeOf(false), v.Type()))
				newIds := append(oldIdsArr, k)
				fl.SetDataIdx(strings.Join(newIds, "."))

				rel, ok := fl.Bool()
				if !ok {
					return false
				}

				m.SetMapIndex(reflect.ValueOf(rel), v)
				fl.SetDataIdx(oldIdx)
			default:
				panic(fmt.Sprintf("uniq valid struct key not support %s", mTyp.Kind()))
			}
		}

		return fl.Len() == m.Len()

	default:
		panic(fmt.Sprintf("Bad field type %s", fl.Raw()))
	}

	return false
}

// isMAC is the validation function for validating if the field's value is a valid MAC address.
func isMAC(fl Field) bool {
	rel, ok := fl.Str()
	if !ok {
		return false
	}

	_, err := net.ParseMAC(rel)

	return err == nil
}

// isIPv4 is the validation function for validating if a value is a valid v4 IP address.
func isIPv4(fl Field) bool {
	rel, ok := fl.Str()
	if !ok {
		return false
	}

	ip := net.ParseIP(rel)

	return ip != nil && ip.To4() != nil
}

// isIPv6 is the validation function for validating if the field's value is a valid v6 IP address.
func isIPv6(fl Field) bool {
	rel, ok := fl.Str()
	if !ok {
		return false
	}

	ip := net.ParseIP(rel)

	return ip != nil && ip.To4() == nil
}

// isIP is the validation function for validating if the field's value is a valid v4 or v6 IP address.
func isIP(fl Field) bool {
	rel, ok := fl.Str()
	if !ok {
		return false
	}

	ip := net.ParseIP(rel)

	return ip != nil
}

// isMD5 is the validation function for validating if the field's value is a valid MD5.
func isMD5(fl Field) bool {
	rel, ok := fl.Str()
	if !ok {
		return false
	}

	return md5Regex.MatchString(rel)
}

// isSHA256 is the validation function for validating if the field's value is a valid SHA256.
func isSHA256(fl Field) bool {
	rel, ok := fl.Str()
	if !ok {
		return false
	}

	return sha256Regex.MatchString(rel)
}

// excludesRune is the validation function for validating that the field's value does not contain the rune specified within the param.
func excludesRune(fl Field) bool {
	return !containsRune(fl)
}

// excludesAll is the validation function for validating that the field's value does not contain any of the characters specified within the param.
func excludesAll(fl Field) bool {
	return !containsAny(fl)
}

// excludes is the validation function for validating that the field's value does not contain the text specified within the param.
func excludes(fl Field) bool {
	return !contains(fl)
}

// containsRune is the validation function for validating that the field's value contains the rune specified within the param.
func containsRune(fl Field) bool {
	r, _ := utf8.DecodeRuneInString(fl.RuleVal())

	rel, ok := fl.Str()
	if !ok {
		return false
	}

	return strings.ContainsRune(rel, r)
}

// containsAny is the validation function for validating that the field's value contains any of the characters specified within the param.
func containsAny(fl Field) bool {
	rel, ok := fl.Str()
	if !ok {
		return false
	}

	return strings.ContainsAny(rel, fl.RuleVal())
}

// contains is the validation function for validating that the field's value contains the text specified within the param.
func contains(fl Field) bool {
	rel, ok := fl.Str()
	if !ok {
		return false
	}

	return strings.Contains(rel, fl.RuleVal())
}

// startsWith is the validation function for validating that the field's value starts with the text specified within the param.
func startsWith(fl Field) bool {
	rel, ok := fl.Str()
	if !ok {
		return false
	}

	return strings.HasPrefix(rel, fl.RuleVal())
}

// endsWith is the validation function for validating that the field's value ends with the text specified within the param.
func endsWith(fl Field) bool {
	rel, ok := fl.Str()
	if !ok {
		return false
	}

	return strings.HasSuffix(rel, fl.RuleVal())
}

// startsNotWith is the validation function for validating that the field's value does not start with the text specified within the param.
func startsNotWith(fl Field) bool {
	return !startsWith(fl)
}

// endsNotWith is the validation function for validating that the field's value does not end with the text specified within the param.
func endsNotWith(fl Field) bool {
	return !endsWith(fl)
}

// isNe is the validation function for validating that the field's value does not equal the provided param value.
func isNe(fl Field) bool {
	return !isEq(fl)
}

// isEqField is the validation function for validating if the current field's value is equal to the field specified by the param's value.
func isEqField(fl Field) bool {
	kind := fl.Self().Kind()

	anotherField, ok := fl.StructField(fl.Parent(), fl.RuleVal())
	if !ok || anotherField.Self().Kind() != kind {
		return false
	}

	// compare field both must has values
	if !fl.Exist() || !anotherField.Exist() {
		return false
	}

	switch kind {

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if fl.IsNumeric() && anotherField.IsNumeric() {
			flv, _ := fl.Int()
			anv, _ := anotherField.Int()

			return flv == anv
		}

		return false

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		if fl.IsNumeric() && anotherField.IsNumeric() {
			flv, _ := fl.Uint()
			anv, _ := anotherField.Uint()

			return flv == anv
		}

		return false

	case reflect.Float32, reflect.Float64:
		if fl.IsNumeric() && anotherField.IsNumeric() {
			flv, _ := fl.Float()
			anv, _ := anotherField.Float()

			return flv == anv
		}

		return false

	case reflect.Slice, reflect.Map, reflect.Array:
		if fl.IsArray() && anotherField.IsArray() {
			return fl.Len() == anotherField.Len()
		}

		if fl.IsObject() && anotherField.IsObject() {
			return fl.Len() == anotherField.Len()
		}

		return false

	case reflect.Bool:
		if fl.IsBool() && anotherField.IsBool() {
			flv, _ := fl.Bool()
			anv, _ := anotherField.Bool()

			return flv == anv
		}

		return false

	case reflect.Struct: // struct only support time.Time{}
		fieldType := fl.Self()

		rel, isOk := fl.Str()
		if !isOk {
			return false
		}

		// Not Same underlying type i.e. struct and time
		if fieldType != anotherField.Self() {
			return false
		}

		if fieldType.ConvertibleTo(TimeType) && anotherField.Self().ConvertibleTo(TimeType) {
			valT, e := FieldTimeVal(rel, fl.Tag())
			if e != nil {
				return false
			}

			currentT, e := FieldTimeVal(anotherField.MustStr(), anotherField.Tag())
			if e != nil {
				return false
			}

			return valT.Equal(currentT)
		}

	}

	// default reflect.DataStr:
	return fl.Raw() == anotherField.Raw()
}

// isEq is the validation function for validating if the current field's value is equal to the param's value.
func isEq(fl Field) bool {
	if !fl.Exist() {
		return false
	}

	switch fl.Self().Kind() {

	case reflect.String:
		rel, ok := fl.Str()
		if !ok {
			return false
		}

		return rel == fl.RuleVal()

	case reflect.Slice, reflect.Map, reflect.Array:
		p, e := strconv.Atoi(fl.RuleVal())
		if e != nil {
			return false
		}

		if fl.IsArray() || fl.IsObject() {
			return int64(fl.Len()) == int64(p)
		}

		return false

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		p, e := strconv.Atoi(fl.RuleVal())
		if e != nil {
			return false
		}

		rel, ok := fl.Int()
		if !ok {
			return false
		}

		return rel == int64(p)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		p, e := strconv.Atoi(fl.RuleVal())
		if e != nil {
			return false
		}

		rel, ok := fl.Uint()
		if !ok {
			return false
		}

		return rel == uint64(p)

	case reflect.Float32, reflect.Float64:
		p, e := strconv.ParseFloat(fl.RuleVal(), 64)
		if e != nil {
			return false
		}

		rel, ok := fl.Float()
		if !ok {
			return false
		}

		return rel == p

	case reflect.Bool:
		p, e := strconv.ParseBool(fl.RuleVal())
		if e != nil {
			return false
		}

		rel, ok := fl.Bool()
		if !ok {
			return false
		}

		return p == rel
	}

	panic(fmt.Sprintf("Bad field type %s", fl.Raw()))
}

// isEqCrossStructField is the validation function for validating that the current field's value is equal to the field, within a separate struct, specified by the param's value.
func isEqCrossStructField(fl Field) bool {
	field := fl.Self()
	kind := field.Kind()

	anotherField, ok := fl.StructField(fl.Root(), fl.RuleVal())
	if !ok || anotherField.Self().Kind() != kind {
		return false
	}

	switch kind {

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		flv, _ := fl.Int()
		anotherV, _ := anotherField.Int()

		return flv == anotherV

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		flv, _ := fl.Uint()
		anotherV, _ := anotherField.Uint()

		return flv == anotherV

	case reflect.Float32, reflect.Float64:
		flv, _ := fl.Float()
		anotherV, _ := anotherField.Float()

		return flv == anotherV

	case reflect.Slice, reflect.Map, reflect.Array:

		return int64(fl.Len()) == int64(anotherField.Len())

	case reflect.Bool:
		flv, _ := fl.Bool()
		anotherV, _ := anotherField.Bool()

		return flv == anotherV

	case reflect.Struct:

		fieldType := fl.Self()
		if fieldType.ConvertibleTo(TimeType) && anotherField.Self().ConvertibleTo(TimeType) {
			flt, flErr := FieldTimeVal(fl.MustStr(), fl.Tag())
			ant, aErr := FieldTimeVal(anotherField.MustStr(), anotherField.Tag())
			if flErr == nil && aErr == nil {
				return flt.Equal(ant)
			}

			if flErr != nil && aErr != nil {
				return fl.Raw() == anotherField.Raw()
			}

			return false
		}

		// Not Same underlying type i.e. struct and time
		if fieldType != anotherField.Self() {
			return false
		}
	}

	// default reflect.DataStr:
	return fl.Raw() == anotherField.Raw()
}

// isNeCrossStructField is the validation function for validating that the current field's value is not equal to the field, within a separate struct, specified by the param's value.
func isNeCrossStructField(fl Field) bool {
	field := fl.Self()
	kind := field.Kind()

	anotherField, ok := fl.StructField(fl.Root(), fl.RuleVal())
	if !ok || anotherField.Self().Kind() != kind {
		return true
	}

	switch kind {

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		flv, _ := fl.Int()
		anotherV, _ := anotherField.Int()

		return flv != anotherV

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		flv, _ := fl.Uint()
		anotherV, _ := anotherField.Uint()

		return flv != anotherV

	case reflect.Float32, reflect.Float64:
		flv, _ := fl.Float()
		anotherV, _ := anotherField.Float()

		return flv != anotherV

	case reflect.Slice, reflect.Map, reflect.Array:

		return int64(fl.Len()) != int64(anotherField.Len())

	case reflect.Bool:
		flv, _ := fl.Bool()
		anotherV, _ := anotherField.Bool()

		return flv != anotherV

	case reflect.Struct:

		fieldType := fl.Self()
		if fieldType.ConvertibleTo(TimeType) && anotherField.Self().ConvertibleTo(TimeType) {
			flt, flErr := FieldTimeVal(fl.MustStr(), fl.Tag())
			ant, anErr := FieldTimeVal(anotherField.MustStr(), anotherField.Tag())
			if flErr == nil && anErr == nil {
				return !flt.Equal(ant)
			}

			// if both parse time err, compare with raw string
			if flErr != nil && anErr != nil {
				return fl.Raw() != anotherField.Raw()
			}

			return true
		}

		// Not Same underlying type i.e. struct and time
		if fieldType != anotherField.Self() {
			return true
		}
	}

	// default reflect.DataStr:
	return fl.Raw() != anotherField.Raw()
}

// isGtCrossStructField is the validation function for validating if the current field's value is greater than the field, within a separate struct, specified by the param's value.
func isGtCrossStructField(fl Field) bool {
	field := fl.Self()
	kind := field.Kind()

	anotherField, ok := fl.StructField(fl.Root(), fl.RuleVal())
	if !ok || fl.Self().Kind() != kind {
		return false
	}

	switch kind {

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		flv, _ := fl.Int()
		anotherV, _ := anotherField.Int()

		return flv > anotherV

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		flv, _ := fl.Uint()
		anotherV, _ := anotherField.Uint()

		return flv > anotherV

	case reflect.Float32, reflect.Float64:
		flv, _ := fl.Float()
		anotherV, _ := anotherField.Float()

		return flv > anotherV

	case reflect.Slice, reflect.Map, reflect.Array:

		return int64(fl.Len()) > int64(anotherField.Len())

	case reflect.Struct:

		fieldType := fl.Self()
		if fieldType.ConvertibleTo(TimeType) && anotherField.Self().ConvertibleTo(TimeType) {
			flt, flErr := FieldTimeVal(fl.MustStr(), fl.Tag())
			ant, anErr := FieldTimeVal(anotherField.MustStr(), anotherField.Tag())
			if flErr == nil && anErr == nil {
				return flt.After(ant)
			}

			// if both parse time err, compare with raw string
			if flErr != nil && anErr != nil {
				return fl.Raw() > anotherField.Raw()
			}

			return false
		}

		// Not Same underlying type i.e. struct and time
		if fieldType != anotherField.Self() {
			return false
		}
	}

	// default reflect.DataStr:
	return fl.Raw() > anotherField.Raw()
}

// isGteCrossStructField is the validation function for validating if the current field's value is greater than the field, within a separate struct, specified by the param's value.
func isGteCrossStructField(fl Field) bool {
	field := fl.Self()
	kind := field.Kind()

	anotherField, ok := fl.StructField(fl.Root(), fl.RuleVal())
	if !ok || fl.Self().Kind() != kind {
		return false
	}

	switch kind {

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		flv, _ := fl.Int()
		anotherV, _ := anotherField.Int()

		return flv >= anotherV

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		flv, _ := fl.Uint()
		anotherV, _ := anotherField.Uint()

		return flv >= anotherV

	case reflect.Float32, reflect.Float64:
		flv, _ := fl.Float()
		anotherV, _ := anotherField.Float()

		return flv >= anotherV

	case reflect.Slice, reflect.Map, reflect.Array:

		return int64(fl.Len()) >= int64(anotherField.Len())

	case reflect.Struct:

		fieldType := fl.Self()
		if fieldType.ConvertibleTo(TimeType) && anotherField.Self().ConvertibleTo(TimeType) {
			flt, flErr := FieldTimeVal(fl.MustStr(), fl.Tag())
			ant, anErr := FieldTimeVal(anotherField.MustStr(), anotherField.Tag())
			if flErr == nil && anErr == nil {
				return flt.After(ant) || flt.Equal(ant)
			}

			// if both parse time err, compare with raw string
			if flErr != nil && anErr != nil {
				return fl.Raw() >= anotherField.Raw()
			}

			return false
		}

		// Not Same underlying type i.e. struct and time
		if fieldType != anotherField.Self() {
			return false
		}
	}

	// default reflect.DataStr:
	return fl.Raw() >= anotherField.Raw()
}

// isLteCrossStructField is the validation function for validating if the current field's value is greater than the field, within a separate struct, specified by the param's value.
func isLteCrossStructField(fl Field) bool {
	field := fl.Self()
	kind := field.Kind()

	anotherField, ok := fl.StructField(fl.Root(), fl.RuleVal())
	if !ok || fl.Self().Kind() != kind {
		return false
	}

	switch kind {

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		flv, _ := fl.Int()
		anotherV, _ := anotherField.Int()

		return flv <= anotherV

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		flv, _ := fl.Uint()
		anotherV, _ := anotherField.Uint()

		return flv <= anotherV

	case reflect.Float32, reflect.Float64:
		flv, _ := fl.Float()
		anotherV, _ := anotherField.Float()

		return flv <= anotherV

	case reflect.Slice, reflect.Map, reflect.Array:

		return int64(fl.Len()) <= int64(anotherField.Len())

	case reflect.Struct:

		fieldType := fl.Self()
		if fieldType.ConvertibleTo(TimeType) && anotherField.Self().ConvertibleTo(TimeType) {
			flt, flErr := FieldTimeVal(fl.MustStr(), fl.Tag())
			ant, anErr := FieldTimeVal(anotherField.MustStr(), anotherField.Tag())
			if flErr == nil && anErr == nil {
				return flt.Before(ant) || flt.Equal(ant)
			}

			// if both parse time err, compare with raw string
			if flErr != nil && anErr != nil {
				return fl.Raw() <= anotherField.Raw()
			}

			return false
		}

		// Not Same underlying type i.e. struct and time
		if fieldType != anotherField.Self() {
			return false
		}
	}

	// default reflect.DataStr:
	return fl.Raw() <= anotherField.Raw()
}

// fieldContains is the validation function for validating if the current field's value contains the field specified by the param's value.
func fieldContains(fl Field) bool {
	anotherField, ok := fl.StructField(fl.Parent(), fl.RuleVal())
	if !ok {
		return false
	}

	if !fl.Exist() || !anotherField.Exist() {
		return false
	}

	return strings.Contains(fl.MustStr(), anotherField.MustStr())
}

// fieldExcludes is the validation function for validating if the current field's value excludes the field specified by the param's value.
func fieldExcludes(fl Field) bool {
	anotherField, ok := fl.StructField(fl.Parent(), fl.RuleVal())
	if !ok {
		return true
	}

	if !fl.Exist() || !anotherField.Exist() {
		return false
	}

	return !strings.Contains(fl.MustStr(), anotherField.MustStr())
}

// isNeField is the validation function for validating if the current field's value is not equal to the field specified by the param's value.
func isNeField(fl Field) bool {
	field := fl.Self()
	kind := field.Kind()

	anotherField, ok := fl.StructField(fl.Parent(), fl.RuleVal())

	if !ok || anotherField.Self().Kind() != kind {
		return true
	}

	switch kind {

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		flv, _ := fl.Int()
		anotherV, _ := anotherField.Int()

		return flv != anotherV

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		flv, _ := fl.Uint()
		anotherV, _ := anotherField.Uint()

		return flv != anotherV

	case reflect.Float32, reflect.Float64:
		flv, _ := fl.Float()
		anotherV, _ := anotherField.Float()

		return flv != anotherV

	case reflect.Slice, reflect.Map, reflect.Array:

		return int64(fl.Len()) != int64(anotherField.Len())

	case reflect.Bool:
		flv, _ := fl.Bool()
		anotherV, _ := anotherField.Bool()

		return flv != anotherV

	case reflect.Struct:

		fieldType := fl.Self()
		if fieldType.ConvertibleTo(TimeType) && anotherField.Self().ConvertibleTo(TimeType) {
			flt, flErr := FieldTimeVal(fl.MustStr(), fl.Tag())
			ant, anErr := FieldTimeVal(anotherField.MustStr(), anotherField.Tag())
			if flErr == nil && anErr == nil {
				return !flt.Equal(ant)
			}

			// if both parse time err, compare with raw string
			if flErr != nil && anErr != nil {
				return fl.Raw() != anotherField.Raw()
			}

			return true
		}

		// Not Same underlying type i.e. struct and time
		if fieldType != anotherField.Self() {
			return true
		}
	}

	// default reflect.DataStr:
	return fl.Raw() != anotherField.Raw()
}

// isGteField is the validation function for validating if the current field's value is greater than or equal to the field specified by the param's value.
func isGteField(fl Field) bool {
	field := fl.Self()
	kind := field.Kind()

	anotherField, ok := fl.StructField(fl.Parent(), fl.RuleVal())
	if !ok || fl.Self().Kind() != kind {
		return false
	}

	switch kind {

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		flv, _ := fl.Int()
		anotherV, _ := anotherField.Int()

		return flv >= anotherV

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		flv, _ := fl.Uint()
		anotherV, _ := anotherField.Uint()

		return flv >= anotherV

	case reflect.Float32, reflect.Float64:
		flv, _ := fl.Float()
		anotherV, _ := anotherField.Float()

		return flv >= anotherV

	case reflect.Slice, reflect.Map, reflect.Array:

		return int64(fl.Len()) >= int64(anotherField.Len())

	case reflect.Struct:

		fieldType := fl.Self()
		if fieldType.ConvertibleTo(TimeType) && anotherField.Self().ConvertibleTo(TimeType) {
			flt, flErr := FieldTimeVal(fl.MustStr(), fl.Tag())
			ant, anErr := FieldTimeVal(anotherField.MustStr(), anotherField.Tag())
			if flErr == nil && anErr == nil {
				return flt.After(ant) || flt.Equal(ant)
			}

			// if both parse time err, compare with raw string
			if flErr != nil && anErr != nil {
				return fl.Raw() >= anotherField.Raw()
			}

			return false
		}

		// Not Same underlying type i.e. struct and time
		if fieldType != anotherField.Self() {
			return false
		}
	}

	// default reflect.DataStr:
	return fl.Raw() >= anotherField.Raw()
}

// isGtField is the validation function for validating if the current field's value is greater than the field specified by the param's value.
func isGtField(fl Field) bool {
	field := fl.Self()
	kind := field.Kind()

	anotherField, ok := fl.StructField(fl.Parent(), fl.RuleVal())
	if !ok || fl.Self().Kind() != kind {
		return false
	}

	switch kind {

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		flv, _ := fl.Int()
		anotherV, _ := anotherField.Int()

		return flv > anotherV

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		flv, _ := fl.Uint()
		anotherV, _ := anotherField.Uint()

		return flv > anotherV

	case reflect.Float32, reflect.Float64:
		flv, _ := fl.Float()
		anotherV, _ := anotherField.Float()

		return flv > anotherV

	case reflect.Slice, reflect.Map, reflect.Array:

		return int64(fl.Len()) > int64(anotherField.Len())

	case reflect.Struct:

		fieldType := fl.Self()
		if fieldType.ConvertibleTo(TimeType) && anotherField.Self().ConvertibleTo(TimeType) {
			flt, flErr := FieldTimeVal(fl.MustStr(), fl.Tag())
			ant, anErr := FieldTimeVal(anotherField.MustStr(), anotherField.Tag())
			if flErr == nil && anErr == nil {
				return flt.After(ant)
			}

			// if both parse time err, compare with raw string
			if flErr != nil && anErr != nil {
				return fl.Raw() > anotherField.Raw()
			}

			return false
		}

		// Not Same underlying type i.e. struct and time
		if fieldType != anotherField.Self() {
			return false
		}
	}

	// default reflect.DataStr:
	return fl.Raw() > anotherField.Raw()
}

// isLteField is the validation function for validating if the current field's value is less than or equal to the field specified by the param's value.
func isLteField(fl Field) bool {
	field := fl.Self()
	kind := field.Kind()

	anotherField, ok := fl.StructField(fl.Parent(), fl.RuleVal())
	if !ok || fl.Self().Kind() != kind {
		return false
	}

	switch kind {

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		flv, _ := fl.Int()
		anotherV, _ := anotherField.Int()

		return flv <= anotherV

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		flv, _ := fl.Uint()
		anotherV, _ := anotherField.Uint()

		return flv <= anotherV

	case reflect.Float32, reflect.Float64:
		flv, _ := fl.Float()
		anotherV, _ := anotherField.Float()

		return flv <= anotherV

	case reflect.Slice, reflect.Map, reflect.Array:

		return int64(fl.Len()) <= int64(anotherField.Len())

	case reflect.Struct:

		fieldType := fl.Self()
		if fieldType.ConvertibleTo(TimeType) && anotherField.Self().ConvertibleTo(TimeType) {
			flt, flErr := FieldTimeVal(fl.MustStr(), fl.Tag())
			ant, anErr := FieldTimeVal(anotherField.MustStr(), anotherField.Tag())
			if flErr == nil && anErr == nil {
				return flt.Before(ant) || flt.Equal(ant)
			}

			// if both parse time err, compare with raw string
			if flErr != nil && anErr != nil {
				return fl.Raw() <= anotherField.Raw()
			}

			return false
		}

		// Not Same underlying type i.e. struct and time
		if fieldType != anotherField.Self() {
			return false
		}
	}

	// default reflect.DataStr:
	return fl.Raw() <= anotherField.Raw()
}

// isLtField is the validation function for validating if the current field's value is less than the field specified by the param's value.
func isLtField(fl Field) bool {
	field := fl.Self()
	kind := field.Kind()

	anotherField, ok := fl.StructField(fl.Parent(), fl.RuleVal())
	if !ok || fl.Self().Kind() != kind {
		return false
	}

	switch kind {

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		flv, _ := fl.Int()
		anotherV, _ := anotherField.Int()

		return flv < anotherV

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		flv, _ := fl.Uint()
		anotherV, _ := anotherField.Uint()

		return flv < anotherV

	case reflect.Float32, reflect.Float64:
		flv, _ := fl.Float()
		anotherV, _ := anotherField.Float()

		return flv < anotherV

	case reflect.Slice, reflect.Map, reflect.Array:

		return int64(fl.Len()) < int64(anotherField.Len())

	case reflect.Struct:

		fieldType := fl.Self()
		if fieldType.ConvertibleTo(TimeType) && anotherField.Self().ConvertibleTo(TimeType) {
			flt, flErr := FieldTimeVal(fl.MustStr(), fl.Tag())
			ant, anErr := FieldTimeVal(anotherField.MustStr(), anotherField.Tag())
			if flErr == nil && anErr == nil {
				return flt.Before(ant)
			}

			// if both parse time err, compare with raw string
			if flErr != nil && anErr != nil {
				return fl.Raw() < anotherField.Raw()
			}

			return false
		}

		// Not Same underlying type i.e. struct and time
		if fieldType != anotherField.Self() {
			return false
		}
	}

	// default reflect.DataStr:
	return fl.Raw() < anotherField.Raw()
}

// isLtCrossStructField is the validation function for validating if the current field's value is greater than the field, within a separate struct, specified by the param's value.
func isLtCrossStructField(fl Field) bool {
	field := fl.Self()
	kind := field.Kind()

	anotherField, ok := fl.StructField(fl.Root(), fl.RuleVal())
	if !ok || fl.Self().Kind() != kind {
		return false
	}

	switch kind {

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		flv, _ := fl.Int()
		anotherV, _ := anotherField.Int()

		return flv < anotherV

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		flv, _ := fl.Uint()
		anotherV, _ := anotherField.Uint()

		return flv < anotherV

	case reflect.Float32, reflect.Float64:
		flv, _ := fl.Float()
		anotherV, _ := anotherField.Float()

		return flv < anotherV

	case reflect.Slice, reflect.Map, reflect.Array:

		return int64(fl.Len()) < int64(anotherField.Len())

	case reflect.Struct:

		fieldType := fl.Self()
		if fieldType.ConvertibleTo(TimeType) && anotherField.Self().ConvertibleTo(TimeType) {
			flt, flErr := FieldTimeVal(fl.MustStr(), fl.Tag())
			ant, anErr := FieldTimeVal(anotherField.MustStr(), anotherField.Tag())
			if flErr == nil && anErr == nil {
				return flt.Before(ant)
			}

			// if both parse time err, compare with raw string
			if flErr != nil && anErr != nil {
				return fl.Raw() < anotherField.Raw()
			}

			return false
		}

		// Not Same underlying type i.e. struct and time
		if fieldType != anotherField.Self() {
			return false
		}
	}

	// default reflect.DataStr:
	return fl.Raw() < anotherField.Raw()
}

// isBase64 is the validation function for validating if the current field's value is a valid base 64.
func isBase64(fl Field) bool {
	rel, ok := fl.Str()
	if !ok {
		return false
	}

	return base64Regex.MatchString(rel)
}

// isBase64URL is the validation function for validating if the current field's value is a valid base64 URL safe string.
func isBase64URL(fl Field) bool {
	rel, ok := fl.Str()
	if !ok {
		return false
	}

	return base64URLRegex.MatchString(rel)
}

// isURI is the validation function for validating if the current field's value is a valid URI.
func isURI(fl Field) bool {
	rel, ok := fl.Str()
	if !ok {
		return false
	}

	switch fl.Self().Kind() {
	case reflect.String:

		s := rel

		// checks needed as of Go 1.6 because of change https://github.com/golang/go/commit/617c93ce740c3c3cc28cdd1a0d712be183d0b328#diff-6c2d018290e298803c0c9419d8739885L195
		// emulate browser and strip the '#' suffix prior to validation. see issue-#237
		if i := strings.Index(s, "#"); i > -1 {
			s = s[:i]
		}

		if len(s) == 0 {
			return false
		}

		_, err := url.ParseRequestURI(s)

		return err == nil
	}

	panic(fmt.Sprintf("Bad field type %s", fl.Raw()))
}

// isURL is the validation function for validating if the current field's value is a valid URL.
func isURL(fl Field) bool {
	rel, ok := fl.Str()
	if !ok {
		return false
	}

	switch fl.Self().Kind() {
	case reflect.String:

		var i int
		s := rel

		// checks needed as of Go 1.6 because of change https://github.com/golang/go/commit/617c93ce740c3c3cc28cdd1a0d712be183d0b328#diff-6c2d018290e298803c0c9419d8739885L195
		// emulate browser and strip the '#' suffix prior to validation. see issue-#237
		if i = strings.Index(s, "#"); i > -1 {
			s = s[:i]
		}

		if len(s) == 0 {
			return false
		}

		urlData, err := url.ParseRequestURI(s)

		if err != nil || urlData.Scheme == "" {
			return false
		}

		return true
	}

	panic(fmt.Sprintf("Bad field type %s", fl.Raw()))
}

// isUrnRFC2141 is the validation function for validating if the current field's value is a valid URN as per RFC 2141.
func isUrnRFC2141(fl Field) bool {
	rel, ok := fl.Str()
	if !ok {
		return false
	}

	switch fl.Self().Kind() {
	case reflect.String:
		_, match := urn.Parse([]byte(rel))
		return match
	}

	panic(fmt.Sprintf("Bad field type %s", fl.Raw()))
}

// isFile is the validation function for validating if the current field's value is a valid file path.
func isFile(fl Field) bool {
	rel, ok := fl.Str()
	if !ok {
		return false
	}

	switch fl.Self().Kind() {
	case reflect.String:
		fileInfo, err := os.Stat(rel)
		if err != nil {
			return false
		}

		return !fileInfo.IsDir()
	}

	panic(fmt.Sprintf("Bad field type %T", fl.Raw()))
}

// isEmail is the validation function for validating if the current field's value is a valid email address.
func isEmail(fl Field) bool {
	rel, ok := fl.Str()
	if !ok {
		return false
	}

	return emailRegex.MatchString(rel)
}

// isNumber valid the value whether only 0-9 number
func isNumber(fl Field) bool {
	if !fl.Exist() {
		return false
	}

	rel, ok := fl.Int()
	if !ok {
		return false
	}

	return numberRegex.MatchString(strconv.FormatInt(rel, 10))
}

// isNumeric check the value whether number , include int、float
func isNumeric(fl Field) bool {
	if !fl.Exist() {
		return false
	}

	return fl.IsNumeric()
}

// isAlphanum is the validation function for validating if the current field's value is a valid alphanumeric value.
func isAlphanum(fl Field) bool {
	if !fl.Exist() {
		return false
	}

	return alphaNumericRegex.MatchString(fl.MustStr())
}

// isAlpha is the validation function for validating if the current field's value is a valid alpha value.
func isAlpha(fl Field) bool {
	if !fl.Exist() {
		return false
	}

	return alphaRegex.MatchString(fl.MustStr())
}

// isAlphanumUnicode is the validation function for validating if the current field's value is a valid alphanumeric unicode value.
func isAlphanumUnicode(fl Field) bool {
	if !fl.Exist() {
		return false
	}

	return alphaUnicodeNumericRegex.MatchString(fl.MustStr())
}

// isAlphaUnicode is the validation function for validating if the current field's value is a valid alpha unicode value.
func isAlphaUnicode(fl Field) bool {
	if !fl.Exist() {
		return false
	}

	return alphaUnicodeRegex.MatchString(fl.MustStr())
}

// isBoolean is the validation function for validating if the current field's value is a valid boolean value or can be safely converted to a boolean value.
func isBoolean(fl Field) bool {
	if !fl.Exist() {
		return false
	}

	return fl.IsBool()
}

func getCurRuleIntVal(fl Field) int {
	p, err := strconv.Atoi(fl.RuleVal())
	if err != nil {
		panic(fmt.Errorf("the Key %s valid rule %s value %s parse err: %s", fl.Key(), fl.RuleName(), fl.RuleVal(), err.Error()))
	}

	return p
}

// isGt is the validation function for validating if the current field's value is greater than the param's value.
func isGt(fl Field) bool {
	if !fl.Exist() { // 数据字段不存在直接返回
		return false
	}

	switch fl.Self().Kind() { // 定义的 Struct 想要的数据类型
	case reflect.String:
		p := getCurRuleIntVal(fl)

		rel, ok := fl.Str()
		if !ok {
			return false
		}

		return int64(utf8.RuneCountInString(rel)) > int64(p)

	case reflect.Slice, reflect.Map, reflect.Array:
		p := getCurRuleIntVal(fl)

		if fl.IsArray() || fl.IsObject() {
			return int64(fl.Len()) > int64(p)
		}

		return false

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		p := getCurRuleIntVal(fl)

		rel, ok := fl.Int()
		if !ok {
			return false
		}

		return rel > int64(p)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		p := getCurRuleIntVal(fl)

		rel, ok := fl.Uint()
		if !ok {
			return false
		}

		return rel > uint64(p)

	case reflect.Float32, reflect.Float64:
		p := getCurRuleIntVal(fl)

		rel, ok := fl.Float()
		if !ok {
			return false
		}

		return rel > float64(p)

	case reflect.Struct:
		rel, ok := fl.Str()
		if !ok {
			return false
		}

		if fl.Self().ConvertibleTo(TimeType) {
			now := time.Now().UTC()
			t, e := FieldTimeVal(rel, fl.Tag())
			if e != nil {
				return false
			}

			// 此时，p 值没参与，没有任何用处
			return t.After(now)
		}
	}

	panic(fmt.Sprintf("Bad field type %s", fl.Raw()))
}

// hasLengthOf is the validation function for validating if the current field's value is equal to the param's value.
func hasLengthOf(fl Field) bool {
	ruleVal, _ := strconv.Atoi(fl.RuleVal())
	if !fl.Exist() {
		return false
	}

	switch fl.Self().Kind() {
	case reflect.String:
		rel, ok := fl.Str()
		if !ok {
			return false
		}

		return int64(utf8.RuneCountInString(rel)) == int64(ruleVal)

	case reflect.Slice, reflect.Map, reflect.Array:
		if fl.IsArray() || fl.IsObject() {
			return fl.Len() == ruleVal
		}

		return false

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		rel, ok := fl.Int()
		if !ok {
			return false
		}

		return rel == int64(ruleVal)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		rel, ok := fl.Uint()
		if !ok {
			return false
		}

		return rel == uint64(ruleVal)

	case reflect.Float32, reflect.Float64:
		rel, ok := fl.Float()
		if !ok {
			return false
		}

		return rel == float64(ruleVal)
	}

	panic(fmt.Sprintf("Bad field type %s", fl.Raw()))
}

// isLte is the validation function for validating if the current field's value is less than or equal to the param's value.
func isLte(fl Field) bool {
	if !fl.Exist() {
		return false
	}

	switch fl.Self().Kind() {
	case reflect.String:
		p := getRuleIntVal(fl)

		rel, ok := fl.Str()
		if !ok {
			return false
		}

		return int64(utf8.RuneCountInString(rel)) <= int64(p)

	case reflect.Slice, reflect.Map, reflect.Array:
		p := getRuleIntVal(fl)

		if fl.IsArray() || fl.IsObject() {
			return int64(fl.Len()) <= int64(p)
		}

		return false

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		p := getRuleIntVal(fl)

		rel, ok := fl.Int()
		if !ok {
			return false
		}

		return rel <= int64(p)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		p := getRuleIntVal(fl)

		rel, ok := fl.Uint()
		if !ok {
			return false
		}

		return rel <= uint64(p)

	case reflect.Float32, reflect.Float64:
		p := getRuleIntVal(fl)

		rel, ok := fl.Float()
		if !ok {
			return false
		}

		return rel <= float64(p)

	case reflect.Struct:
		rel, ok := fl.Str()
		if !ok {
			return false
		}

		if fl.Self().ConvertibleTo(TimeType) {
			now := time.Now().UTC()
			t, e := FieldTimeVal(rel, fl.Tag())
			if e != nil {
				return false
			}

			// 此时，p 值没参与，没有任何用处
			return t.Before(now) || t.Equal(now)
		}
	}

	panic(fmt.Sprintf("Bad field type %s", fl.Self().Kind().String()))

}

// isLt is the validation function for validating if the current field's value is less than the param's value.
func isLt(fl Field) bool {
	if !fl.Exist() {
		return false
	}

	switch fl.Self().Kind() { // 定义的 Struct 想要的数据类型
	case reflect.String:
		p := getRuleIntVal(fl)

		rel, ok := fl.Str()
		if !ok {
			return false
		}

		return int64(utf8.RuneCountInString(rel)) < int64(p)

	case reflect.Slice, reflect.Map, reflect.Array:
		p := getRuleIntVal(fl)

		if fl.IsArray() || fl.IsObject() {
			return int64(fl.Len()) < int64(p)
		}

		return false

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		p := getRuleIntVal(fl)

		rel, ok := fl.Int()
		if !ok {
			return false
		}

		return rel < int64(p)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		p := getRuleIntVal(fl)

		rel, ok := fl.Uint()
		if !ok {
			return false
		}

		return rel < uint64(p)

	case reflect.Float32, reflect.Float64:
		p := getRuleIntVal(fl)

		rel, ok := fl.Float()
		if !ok {
			return false
		}

		return rel < float64(p)

	case reflect.Struct:
		rel, ok := fl.Str()
		if !ok {
			return false
		}

		if fl.Self().ConvertibleTo(TimeType) {
			now := time.Now().UTC()
			t, e := FieldTimeVal(rel, fl.Tag())
			if e != nil {
				return false
			}

			// 此时，p 值没参与，没有任何用处
			return t.Before(now)
		}
	}

	panic(fmt.Sprintf("Bad field type %s", fl.Raw()))
}

// hasMaxOf is the validation function for validating if the current field's value is less than or equal to the param's value.
func hasMaxOf(fl Field) bool {
	return isLte(fl)
}

// isTCP4AddrResolvable is the validation function for validating if the field's value is a resolvable tcp4 address.
func isTCP4AddrResolvable(fl Field) bool {
	if !isIP4Addr(fl) {
		return false
	}

	val, ok := fl.Str()
	if !ok {
		return false
	}

	_, err := net.ResolveTCPAddr("tcp4", val)
	return err == nil
}

// isTCP6AddrResolvable is the validation function for validating if the field's value is a resolvable tcp6 address.
func isTCP6AddrResolvable(fl Field) bool {
	if !isIP6Addr(fl) {
		return false
	}

	val, ok := fl.Str()
	if !ok {
		return false
	}

	_, err := net.ResolveTCPAddr("tcp6", val)

	return err == nil
}

// isTCPAddrResolvable is the validation function for validating if the field's value is a resolvable tcp address.
func isTCPAddrResolvable(fl Field) bool {
	if !isIP4Addr(fl) && !isIP6Addr(fl) {
		return false
	}

	val, ok := fl.Str()
	if !ok {
		return false
	}

	_, err := net.ResolveTCPAddr("tcp", val)

	return err == nil
}

// isUDP4AddrResolvable is the validation function for validating if the field's value is a resolvable udp4 address.
func isUDP4AddrResolvable(fl Field) bool {
	if !isIP4Addr(fl) {
		return false
	}

	val, ok := fl.Str()
	if !ok {
		return false
	}

	_, err := net.ResolveUDPAddr("udp4", val)

	return err == nil
}

// isUDP6AddrResolvable is the validation function for validating if the field's value is a resolvable udp6 address.
func isUDP6AddrResolvable(fl Field) bool {
	if !isIP6Addr(fl) {
		return false
	}

	val, ok := fl.Str()
	if !ok {
		return false
	}

	_, err := net.ResolveUDPAddr("udp6", val)

	return err == nil
}

// isUDPAddrResolvable is the validation function for validating if the field's value is a resolvable udp address.
func isUDPAddrResolvable(fl Field) bool {
	if !isIP4Addr(fl) && !isIP6Addr(fl) {
		return false
	}

	val, ok := fl.Str()
	if !ok {
		return false
	}

	_, err := net.ResolveUDPAddr("udp", val)

	return err == nil
}

// isIP4AddrResolvable is the validation function for validating if the field's value is a resolvable ip4 address.
func isIP4AddrResolvable(fl Field) bool {
	if !isIPv4(fl) {
		return false
	}

	val, ok := fl.Str()
	if !ok {
		return false
	}

	_, err := net.ResolveIPAddr("ip4", val)

	return err == nil
}

// isIP6AddrResolvable is the validation function for validating if the field's value is a resolvable ip6 address.
func isIP6AddrResolvable(fl Field) bool {
	if !isIPv6(fl) {
		return false
	}

	val, ok := fl.Str()
	if !ok {
		return false
	}

	_, err := net.ResolveIPAddr("ip6", val)

	return err == nil
}

// isIPAddrResolvable is the validation function for validating if the field's value is a resolvable ip address.
func isIPAddrResolvable(fl Field) bool {
	if !isIP(fl) {
		return false
	}

	val, ok := fl.Str()
	if !ok {
		return false
	}

	_, err := net.ResolveIPAddr("ip", val)

	return err == nil
}

// isUnixAddrResolvable is the validation function for validating if the field's value is a resolvable unix address.
func isUnixAddrResolvable(fl Field) bool {
	val, ok := fl.Str()
	if !ok {
		return false
	}

	_, err := net.ResolveUnixAddr("unix", val)

	return err == nil
}

func isIP4Addr(fl Field) bool {
	val, ok := fl.Str()
	if !ok {
		return false
	}

	if idx := strings.LastIndex(val, ":"); idx != -1 {
		val = val[0:idx]
	}

	ip := net.ParseIP(val)

	return ip != nil && ip.To4() != nil
}

func isIP6Addr(fl Field) bool {
	val, ok := fl.Str()
	if !ok {
		return false
	}

	if idx := strings.LastIndex(val, ":"); idx != -1 {
		if idx != 0 && val[idx-1:idx] == "]" {
			val = val[1 : idx-1]
		}
	}

	ip := net.ParseIP(val)

	return ip != nil && ip.To4() == nil
}

func isHostnameRFC952(fl Field) bool {
	rel, ok := fl.Str()
	if !ok {
		return false
	}

	return hostnameRegexRFC952.MatchString(rel)
}

func isHostnameRFC1123(fl Field) bool {
	rel, ok := fl.Str()
	if !ok {
		return false
	}

	return hostnameRegexRFC1123.MatchString(rel)
}

// isJSON is the validation function for validating if the current field's value is a valid json string.
func isJSON(fl Field) bool {
	rel, ok := fl.Str()
	if !ok {
		return false
	}

	if fl.Self().Kind() == reflect.String {
		return json.Valid([]byte(rel))
	}

	panic(fmt.Sprintf("Bad field type %s", fl.Raw()))
}

// isJWT is the validation function for validating if the current field's value is a valid JWT string.
func isJWT(fl Field) bool {
	rel, ok := fl.Str()
	if !ok {
		return false
	}

	return jWTRegex.MatchString(rel)
}

// isHostnamePort validates a <dns>:<port> combination for fields typically used for socket address.
func isHostnamePort(fl Field) bool {
	rel, ok := fl.Str()
	if !ok {
		return false
	}

	host, port, err := net.SplitHostPort(rel)
	if err != nil {
		return false
	}
	// Port must be int <= 65535.
	if portNum, err := strconv.ParseInt(port, 10, 32); err != nil || portNum > 65535 || portNum < 1 {
		return false
	}

	// If host is specified, it should match a DNS name
	if host != "" {
		return hostnameRegexRFC1123.MatchString(host)
	}
	return true
}

// isLowercase is the validation function for validating if the current field's value is a lowercase string.
func isLowercase(fl Field) bool {
	rel, ok := fl.Str()
	if !ok {
		return false
	}

	if fl.Self().Kind() == reflect.String {
		if rel == "" {
			return false
		}

		return rel == strings.ToLower(rel)
	}

	panic(fmt.Sprintf("Bad field type %s", fl.Raw()))
}

// isUppercase is the validation function for validating if the current field's value is an uppercase string.
func isUppercase(fl Field) bool {
	rel, ok := fl.Str()
	if !ok {
		return false
	}

	if fl.Self().Kind() == reflect.String {
		if rel == "" {
			return false
		}
		return rel == strings.ToUpper(rel)
	}

	panic(fmt.Sprintf("Bad field type %T", fl.Raw()))
}

// isDatetime is the validation function for validating if the current field's value is a valid datetime string.
func isDatetime(fl Field) bool {
	param := fl.RuleVal()

	rel, ok := fl.Str()
	if !ok {
		return false
	}

	if fl.Self().Kind() == reflect.String {
		_, err := time.Parse(param, rel)

		return err == nil
	}

	panic(fmt.Sprintf("Bad field type %s", fl.Raw()))
}

// isTimeZone is the validation function for validating if the current field's value is a valid time zone string.
func isTimeZone(fl Field) bool {
	rel, ok := fl.Str()
	if !ok {
		return false
	}

	if fl.Self().Kind() == reflect.String {
		// empty value is converted to UTC by time.LoadLocation but disallow it as it is not a valid time zone name
		if rel == "" {
			return false
		}

		// Local value is converted to the current system time zone by time.LoadLocation but disallow it as it is not a valid time zone name
		if strings.ToLower(rel) == "local" {
			return false
		}

		_, err := time.LoadLocation(rel)
		return err == nil
	}

	panic(fmt.Sprintf("Bad field type %s", fl.Raw()))
}
