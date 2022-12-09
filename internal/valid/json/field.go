package json

import (
	"github.com/ntt360/gin/internal/valid/rule"
	"github.com/tidwall/gjson"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

const (
	namespaceSeparator = "."
	leftBracket        = "["
	rightBracket       = "]"
)

type Field struct {
	*rule.BaseFiled

	data            gjson.Result  // all json data
	defaultValExist bool          // weather default val exist or not
	defaultVal      string        // default val
	root            reflect.Type  // the struct root field
	parent          reflect.Type  // current parent field
	self            reflect.Type  // want type
	inherit         rule.Inherits // inherit valid rules
	disableValid    bool          // disable valid
}

func (f *Field) Parent() reflect.Type {
	return f.parent
}

func (f *Field) StructField(valTyp reflect.Type, namespace string) (rule.Field, bool) {
	val := reflect.Zero(valTyp)

	// TODO 父级路径和root路径问题
	// 需要准备搜索路径，后续需要通过搜索路径，获取找到元素的所对应的JSON值
	track := append(make([]string, 0, len(f.Track)), f.Track...)
	track = track[0 : len(track)-1] // 定位到父级

	cf := &Field{
		BaseFiled: &rule.BaseFiled{
			CurKey:       valTyp.Name(),
			CurDataIdx:   namespace,
			ValidTagName: f.ValidTagName,
		},
		data: f.data,
		self: valTyp,
	}

BEGIN:
	kind := val.Kind()
	if kind == reflect.Invalid {
		return cf, false
	}

	// recurse loop util namespace is empty
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

		if !val.Type().ConvertibleTo(rule.TimeType) {
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

			param := rule.Param(vTyp.Tag, f.ValidTagName)
			if len(param) == 0 {
				return cf, false
			}

			track = append(track, param)
			val = val.FieldByName(fld)
			namespace = ns

			cf.CurTag = vTyp.Tag
			cf.CurDataIdx = strings.Join(track, ".")
			cf.CurKey = vTyp.Name
			cf.self = vTyp.Type

			// struct key try get the default val
			cf.defaultVal, cf.defaultValExist = rule.DefaultVal(cf.CurTag, cf.ValidTagName)

			goto BEGIN
		}

	case reflect.Array, reflect.Slice:
		idx := strings.Index(namespace, leftBracket)
		idx2 := strings.Index(namespace, rightBracket)

		arrIdx := namespace[idx+1 : idx2]

		startIdx := idx2 + 1
		if startIdx < len(namespace) {
			if namespace[startIdx:startIdx+1] == namespaceSeparator {
				startIdx++
			}
		}

		track = append(track, arrIdx)
		val = reflect.Zero(val.Type().Elem())
		namespace = namespace[startIdx:]

		cf.CurTag = ""
		cf.CurDataIdx = strings.Join(track, ".")
		cf.CurKey = val.Type().Name()
		cf.defaultVal = ""
		cf.defaultValExist = false
		cf.self = val.Type()

		goto BEGIN

	case reflect.Map:
		idx := strings.Index(namespace, leftBracket) + 1
		idx2 := strings.Index(namespace, rightBracket)

		endIdx := idx2
		if endIdx+1 < len(namespace) {
			if namespace[endIdx+1:endIdx+2] == namespaceSeparator {
				endIdx++
			}
		}

		key := namespace[idx:idx2]
		track = append(track, key)
		val = reflect.Zero(val.Type().Elem())

		namespace = namespace[endIdx+1:]

		cf.CurTag = ""
		cf.CurDataIdx = strings.Join(track, ".")
		cf.CurKey = val.Type().Name()
		cf.defaultVal = ""
		cf.defaultValExist = false
		cf.self = val.Type()

		goto BEGIN
	}

	// if got here there was more namespace, cannot go any deeper
	panic("Invalid field namespace")
}

func (f *Field) Root() reflect.Type {
	return f.root
}

func (f *Field) Self() reflect.Type {
	return f.self
}

func (f *Field) Key() string {
	return f.CurParam
}

func (f *Field) RuleName() string {
	return f.CurRule.Name
}

func (f *Field) RuleVal() string {
	return f.CurRule.Val
}

func (f *Field) CurDataIndex() string {
	return f.CurDataIdx
}

func (f *Field) SetDataIdx(idx string) {
	f.CurDataIdx = idx
}

func (f *Field) Tag() reflect.StructTag {
	return f.CurTag
}

func (f *Field) DefaultValExits() bool {
	return f.defaultValExist
}

func (f *Field) DefaultVal() string {
	return f.defaultVal
}

func (f *Field) Exist() bool {
	rel := f.data.Get(f.CurDataIdx).Exists()

	return rel || f.DefaultValExits()
}

func (f *Field) IsBool() bool {
	data := f.data.Get(f.CurDataIdx)
	if !data.Exists() {
		if f.DefaultValExits() {
			rel, e := strconv.ParseBool(f.defaultVal)
			if e != nil {
				return false
			}

			return rel
		}
		return false
	}

	return data.IsBool()
}

func (f *Field) IsArray() bool {
	data := f.data.Get(f.CurDataIdx)
	if !data.Exists() {
		return false
	}

	return data.IsArray()
}

func (f *Field) IsObject() bool {
	data := f.data.Get(f.CurDataIdx)
	if !data.Exists() {
		return false
	}

	return data.IsObject()
}

func (f *Field) IsNumeric() bool {
	data := f.data.Get(f.CurDataIdx)
	if !data.Exists() {
		if !f.DefaultValExits() { // check has default value
			return false
		}

		return regexp.MustCompile(rule.NumericRegexString).MatchString(f.DefaultVal())
	}

	return data.Type == gjson.Number
}

func (f *Field) Str() (string, bool) {
	data := f.data.Get(f.CurDataIdx)
	if !data.Exists() {
		if !f.DefaultValExits() {
			return "", false
		}

		return f.DefaultVal(), true
	}

	if data.Type != gjson.String {
		return "", false
	}

	return data.Str, true
}

func (f *Field) MustStr() string {
	rel, _ := f.Str()

	return rel
}

func (f *Field) Int() (int64, bool) {
	data := f.data.Get(f.CurDataIdx)
	if !data.Exists() {
		if !f.DefaultValExits() {
			return 0, false
		}

		rel, e := strconv.ParseInt(f.DefaultVal(), 10, 64)
		if e != nil {
			return 0, false
		}

		return rel, true
	}

	if data.Type != gjson.Number {
		return 0, false
	}

	return data.Int(), true
}

func (f *Field) InheritTag() reflect.StructTag {
	return f.CurInheritTag
}

func (f *Field) MustInt() int64 {
	rel, _ := f.Int()

	return rel
}

func (f *Field) Uint() (uint64, bool) {
	data := f.data.Get(f.CurDataIdx)
	if !data.Exists() {
		if !f.DefaultValExits() {
			return 0, false
		}

		rel, e := strconv.ParseUint(f.DefaultVal(), 10, 64)
		if e != nil {
			return 0, false
		}

		return rel, true
	}

	if data.Type != gjson.Number {
		return 0, false
	}

	return data.Uint(), true
}

func (f *Field) MustUint() uint64 {
	rel, _ := f.Uint()

	return rel
}

func (f *Field) Float() (float64, bool) {
	data := f.data.Get(f.CurDataIdx)
	if !data.Exists() {
		if !f.DefaultValExits() {
			return 0, false
		}

		rel, e := strconv.ParseFloat(f.DefaultVal(), 64)
		if e != nil {
			return 0, false
		}

		return rel, true
	}

	if data.Type != gjson.Number {
		return 0, false
	}

	return data.Float(), true
}

func (f *Field) MustFloat() float64 {
	rel, _ := f.Float()

	return rel
}

func (f *Field) Len() int {
	data := f.data.Get(f.CurDataIdx)
	if !data.Exists() {
		return 0
	}

	if data.IsArray() {
		return len(data.Array())
	}

	if data.IsObject() {
		return len(data.Map())
	}

	return len(data.Str)
}

func (f *Field) Bool() (bool, bool) {
	data := f.data.Get(f.CurDataIdx)
	if !data.Exists() {
		if !f.DefaultValExits() {
			return false, false
		}
		rel, e := strconv.ParseBool(f.DefaultVal())
		if e != nil {
			return false, false
		}

		return rel, true
	}

	if !data.IsBool() {
		return false, false
	}

	return data.Bool(), true
}

func (f *Field) MustBool() bool {
	rel, _ := f.Bool()

	return rel
}

func (f *Field) Raw() string {
	return f.data.Get(f.CurDataIdx).Raw
}

func (f *Field) ParamKey() string {
	return f.ValidTagName
}

func (f *Field) MapKeys() []string {
	d := f.data.Get(f.CurDataIdx)
	if !d.Exists() {
		return nil
	}

	var rel []string
	for key := range d.Map() {
		rel = append(rel, key)
	}

	return rel
}

func (f *Field) ToString() string {
	d := f.data.Get(f.CurDataIdx)
	if !d.Exists() {
		return ""
	}

	return d.String()
}
