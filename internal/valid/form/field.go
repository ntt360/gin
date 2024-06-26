package form

import (
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/ntt360/gin/internal/valid/rule"
)

const (
	namespaceSeparator = "."
	leftBracket        = "["
	rightBracket       = "]"
)

var timeType = reflect.TypeOf(time.Time{})

type Field struct {
	*rule.BaseFiled

	root            reflect.Type  // root field
	parent          reflect.Type  // parent field
	selfType        reflect.Type  // current field type
	data            Data          // field data
	defaultVal      string        // field default value
	defaultValExist bool          // field default value weather or not
	inherit         rule.Inherits // inherit valid name
	disableValid    bool
}

func (f *Field) Root() reflect.Type {
	return f.root
}

func (f *Field) Parent() reflect.Type {
	return f.parent
}

func (f *Field) StructField(valTyp reflect.Type, namespace string) (rule.Field, bool) {
	val := reflect.Zero(valTyp)

	var track []string

	cf := &Field{
		BaseFiled: &rule.BaseFiled{
			CurKey:       valTyp.Name(),
			CurDataIdx:   namespace,
			ValidTagName: f.ValidTagName,
		},
		root:     f.root,
		selfType: valTyp,
		data:     f.data,
	}

BEGIN:
	kind := val.Kind()
	if kind == reflect.Invalid {
		return nil, false
	}

	if namespace == "" {
		return cf, true
	}

	switch kind {

	case reflect.Interface, reflect.Ptr:
		val = val.Elem()
		valTyp = val.Type()

		goto BEGIN

	case reflect.Struct:
		fld := namespace
		var ns string

		if !val.Type().ConvertibleTo(timeType) {
			idx := strings.Index(namespace, namespaceSeparator)

			if idx != -1 {
				fld = namespace[:idx]
				ns = namespace[idx+1:]
			} else {
				ns = ""
			}

			bracketIdx := strings.Index(fld, leftBracket)
			if bracketIdx != -1 {
				fld = fld[:bracketIdx]
				ns = namespace[bracketIdx:]
			}

			vTyp, ok := val.Type().FieldByName(fld)
			if !ok || vTyp.Anonymous {
				return cf, false
			}

			param := vTyp.Tag.Get(f.ValidTagName)
			if param == "" || param == "-" {
				param = strings.ToLower(vTyp.Name)
			}

			if len(param) == 0 {
				return cf, false
			}

			track = append(track, param)
			val = val.FieldByName(fld)
			namespace = ns

			cf.CurTag = vTyp.Tag
			cf.CurDataIdx = strings.Join(track, ".")
			cf.CurKey = vTyp.Name
			cf.selfType = vTyp.Type

			// struct key try get the default val
			v, vOk := rule.DefaultVal(cf.CurTag, f.ValidTagName)
			cf.defaultVal = v
			cf.defaultValExist = vOk

			goto BEGIN
		}

	case reflect.Array, reflect.Slice:
		idx := strings.Index(namespace, leftBracket)
		idx2 := strings.Index(namespace, rightBracket)

		_ = namespace[idx+1 : idx2]

		startIdx := idx2 + 1
		if startIdx < len(namespace) {
			if namespace[startIdx:startIdx+1] == namespaceSeparator {
				startIdx++
			}
		}
		val = reflect.Zero(val.Type().Elem())
		namespace = namespace[startIdx:]

		goto BEGIN

	case reflect.Map:

		goto BEGIN
	}

	// if got here there was more namespace, cannot go any deeper
	panic("Invalid field namespace")
}

func (f *Field) Key() string {
	return f.CurKey
}

func (f *Field) RuleName() string {
	return f.CurRule.Name
}

func (f *Field) RuleVal() string {
	return f.CurRule.Val
}

func (f *Field) Tag() reflect.StructTag {
	return f.CurTag
}

func (f *Field) InheritTag() reflect.StructTag {
	return f.CurInheritTag
}

func (f *Field) CurDataIndex() string {
	return f.CurDataIdx
}

func (f *Field) DefaultValExits() bool {
	return f.defaultValExist
}

func (f *Field) DefaultVal() string {
	return f.defaultVal
}

func (f *Field) Empty() bool {
	// check default value exist first
	if f.DefaultValExits() {
		return false
	}

	v, ok := f.data.Value(f.CurDataIdx)
	if !ok {
		return true
	}

	// form.Value are always []string
	relVal, _ := v.Interface().([]string)
	for _, s := range relVal {
		if s != "" {
			return false
		}
	}

	return true
}

func (f *Field) Exist() bool {
	_, ok := f.data.Value(f.CurDataIdx)

	return ok || f.DefaultValExits()
}

func (f *Field) IsBool() bool {
	v, ok := f.dataStr()
	if !ok {
		if !f.DefaultValExits() {
			return false
		}

		v = f.DefaultVal()
	}

	rel, e := strconv.ParseBool(v)
	if e != nil {
		return false
	}

	return rel
}

func (f *Field) IsArray() bool {
	v, ok := f.data.Value(f.CurDataIndex())
	if !ok {
		return false
	}

	if v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		v = v.Elem()
	}

	return v.Kind() == reflect.Slice || v.Kind() == reflect.Array
}

func (f *Field) IsObject() bool {
	v, ok := f.data.Value(f.CurDataIndex())
	if !ok {
		return false
	}

	k := v.Kind()
	if k == reflect.Ptr || k == reflect.Interface {
		v = v.Elem()
	}

	return k == reflect.Map || k == reflect.Struct
}

// IsNumeric 判断当前字段数据是否是有效的数字
func (f *Field) IsNumeric() bool {
	v, ok := f.dataStr()
	if !ok {
		if !f.DefaultValExits() {
			return false
		}

		v = f.DefaultVal()
	}

	return regexp.MustCompile(rule.NumericRegexString).MatchString(v)
}

func (f *Field) Str() (string, bool) {
	return f.dataStr()
}

func (f *Field) MustStr() string {
	rel, _ := f.Str()

	return rel
}

// Len get the current index data length
func (f *Field) Len() int {
	val, ok := f.data.Value(f.CurDataIndex())
	if !ok && !f.DefaultValExits() {
		return 0
	}

	if !ok {
		val = reflect.ValueOf(f.DefaultVal())
	}

	if val.Kind() == reflect.Ptr || val.Kind() == reflect.Interface {
		val = val.Elem()
	}

	switch val.Type().Kind() {
	case reflect.Slice, reflect.Array, reflect.Map:
		return val.Len()

	case reflect.String:
		v, vOk := val.Interface().(string)
		if vOk {
			return len(v)
		}

		return 0

	default:
		return 0
	}
}

func (f *Field) dataStr() (string, bool) {
	v, ok := f.data.Value(f.CurDataIdx)
	if !ok {
		if !f.DefaultValExits() {
			return "", false
		}

		return f.DefaultVal(), true
	}

	if !ok {
		v = reflect.ValueOf(f.defaultVal)
	}

	t := v.Kind()

	if t == reflect.Ptr || t == reflect.Interface {
		t = v.Elem().Kind()
		v = v.Elem()
	}

	vStr := ""
	switch t {
	case reflect.Slice, reflect.Array:
		if v.Len() == 1 {
			vArr, vOk := v.Interface().([]string)
			if !vOk {
				return "", false
			}

			vStr = vArr[0]
		}
	case reflect.String:
		vStr, ok = v.Interface().(string)
		if ok {
			return vStr, ok
		}

		return "", false
	}

	return vStr, true
}

func (f *Field) Int() (int64, bool) {
	vStr, ok := f.dataStr()
	if !ok {
		if !f.DefaultValExits() {
			return 0, false
		}

		rel, e := strconv.ParseInt(f.DefaultVal(), 10, 64)
		if e != nil {
			return 0, false
		}

		return rel, true
	}

	intV, e := strconv.ParseInt(vStr, 10, 64)
	if e != nil {
		return 0, false
	}

	return intV, true
}

func (f *Field) MustInt() int64 {
	rel, _ := f.Int()

	return rel
}

func (f *Field) Uint() (uint64, bool) {
	vStr, ok := f.dataStr()
	if !ok {
		if !f.DefaultValExits() {
			return 0, false
		}

		vStr = f.DefaultVal()
	}

	intV, e := strconv.ParseUint(vStr, 10, 64)
	if e != nil {
		return 0, false
	}

	return intV, true
}

func (f *Field) MustUint() uint64 {
	rel, _ := f.Uint()

	return rel
}

func (f *Field) Float() (float64, bool) {
	vStr, ok := f.dataStr()
	if !ok {
		if !f.DefaultValExits() {
			return 0, false
		}
		vStr = f.DefaultVal()
	}

	rel, e := strconv.ParseFloat(vStr, 64)
	if e != nil {
		return 0, false
	}

	return rel, true
}

func (f *Field) MustFloat() float64 {
	rel, _ := f.Float()

	return rel
}

func (f *Field) Raw() string {
	vStr, ok := f.dataStr()
	if ok {
		return vStr
	}

	return ""
}

func (f *Field) SetDataIdx(idx string) {
	f.CurDataIdx = idx
}

func (f *Field) Self() reflect.Type {
	return f.selfType
}

func (f *Field) Bool() (bool, bool) {
	vStr, ok := f.dataStr()
	if !ok {
		if !f.DefaultValExits() {
			return false, false
		}
		vStr = f.DefaultVal()
	}

	rel, e := strconv.ParseBool(vStr)
	if e != nil {
		return false, false
	}

	return rel, true
}

func (f *Field) MustBool() bool {
	rel, _ := f.Bool()

	return rel
}

func (f *Field) ParamKey() string {
	return f.ValidTagName
}

func (f *Field) MapKeys() []string {
	return nil
}

func (f *Field) ToString() string {
	vStr, ok := f.Str()
	if !ok {
		return ""
	}

	return vStr
}
