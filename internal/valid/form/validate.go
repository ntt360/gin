package form

import (
	"fmt"
	"mime/multipart"
	"reflect"
	"strconv"
	"strings"

	"github.com/ntt360/gin/internal/valid/rule"
)

func Validate(data any, m any, validTagName string) error {
	val := reflect.ValueOf(m)
	t := val.Kind()

	if t == reflect.Ptr {
		val = val.Elem()
		t = val.Kind()
	}

	if t != reflect.Struct {
		panic("bind Object only support Struct or *Struct")
	}

	dataV, ok := data.(map[string]any)
	if !ok {
		panic("the data type not form data type")
	}

	f := &Field{
		BaseFiled: &rule.BaseFiled{
			ValidTagName: validTagName,
		},
		root: val.Type(),
		data: dataV,
	}

	// recurse valid
	_, err := recursiveValid(val, f)
	if err != nil {
		return err
	}

	return nil
}

// recurse valid
func recursiveValid(val reflect.Value, field *Field) (bool, error) {
	// current elem type
	t := val.Type().Kind()
	var err error

	// in-depth valid flag
	var deepValid bool

	switch t {
	case reflect.Bool, reflect.String,
		reflect.Uint, reflect.Uint8, reflect.Uint64, reflect.Uint16, reflect.Uint32,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Float32, reflect.Float64:

		if !field.disableValid {
			_, err = valid(field)
			if err != nil {
				return false, err
			}
		}

		err = trySetVal(val, field)
		if err != nil {
			return false, err
		}

	case reflect.Slice, reflect.Array:

		// valid array self
		deepValid, err = valid(field)
		if err != nil {
			return false, err
		}

		l := field.Len()
		tmpArr := reflect.MakeSlice(val.Type(), 0, l)

		// foreach array elements
		for j := 0; j < l; j++ {
			childTyp := val.Type().Elem()
			el := reflect.New(childTyp)

			field.PushTrack(strconv.Itoa(j))
			field.parent = val.Type()
			field.selfType = childTyp
			field.CurInheritTag = field.CurTag
			field.CurTag = ""
			field.disableValid = !deepValid

			unValid, rErr := recursiveValid(el, field)
			if rErr != nil {
				return unValid, rErr
			}

			field.PopTrack()

			// add to tmp array
			tmpArr = reflect.Append(tmpArr, reflect.Indirect(el))
		}

		field.CurInheritTag = ""
		field.inherit = nil
		field.disableValid = false

		// override val array
		val.Set(tmpArr)

	case reflect.Struct: // valid struct field
		field.selfType = val.Type()
		field.CurKey = val.Type().Name()

		// valid struct self
		_, err = valid(field)
		if err != nil {
			return false, err
		}

		// the time.Time struct no need deep in valid
		if val.Type().ConvertibleTo(rule.TimeType) {
			tStr, ok := field.Str()
			if !ok {
				return true, nil
			}

			tVal, e := rule.FieldTimeVal(tStr, field.CurTag)
			if e != nil {
				return false, field.WrapperErr(fmt.Errorf("the param %s not valid", field.CurParam))
			}

			val.Set(reflect.ValueOf(tVal))
			return false, nil
		}

		// multipart files
		if val.Type().ConvertibleTo(reflect.TypeOf(multipart.FileHeader{})) {
			fVal, ok := field.data.Value(field.CurDataIdx)
			if ok {
				switch fVal.Type().Kind() {
				case reflect.Interface:
					fVal = fVal.Elem()
					switch fVal.Interface().(type) {
					case []*multipart.FileHeader:
						val.Set(fVal.Index(0).Elem())
					}
				case reflect.Ptr:
					val.Set(fVal.Elem())
				}
			}
			return false, nil
		}

		// valid struct per key value
		n := val.NumField()
		for i := 0; i < n; i++ {
			structElem := val.Field(i)
			vt := val.Type().Field(i)

			if !vt.IsExported() { //if the key is private, just skipped
				continue
			}

			if !vt.Anonymous {
				field.parent = val.Type()
				field.selfType = vt.Type
				field.CurKey = vt.Name
				field.CurTag = vt.Tag
				field.CurInheritTag = ""

				d, dOk := rule.DefaultVal(field.CurTag, field.ValidTagName)
				field.defaultVal = d
				field.defaultValExist = dOk
				field.CurParam = rule.Param(field.CurTag, field.ValidTagName)
				field.PushTrack(field.CurParam)
			}

			unValid, rErr := recursiveValid(structElem, field)
			if rErr != nil {
				return unValid, rErr
			}

			field.PopTrack()
		}

	case reflect.Ptr: // give the pointer real elem
		if !val.Elem().IsValid() { // val child maybe is zero value or pointer value, such as *int
			v := reflect.New(val.Type().Elem())
			field.selfType = v.Type()
			unValid, rErr := recursiveValid(v.Elem(), field)
			if rErr != nil {
				return unValid, rErr
			}

			if !unValid {
				val.Set(v)
			}

			return false, nil
		} else {
			val = val.Elem()
		}

		unValid, rErr := recursiveValid(val, field)
		if rErr != nil {
			return unValid, rErr
		}
	case reflect.Interface: // interface data not valid
		v, ok := field.data.Value(field.CurDataIndex())
		if !ok {
			// TODO default value
			return false, nil
		}

		val.Set(v)
		return false, nil
	case reflect.Map: // form data not support map type
		return false, nil
	}

	return false, err
}

func trySetVal(val reflect.Value, f *Field) (e error) {
	if !f.Exist() {
		return nil
	}

	switch val.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		rel, ok := f.Int()
		if !ok {
			return f.WrapperErr(fmt.Errorf("the param %s data type is not int", f.CurParam))
		}

		val.SetInt(rel)

	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint, reflect.Uintptr:
		rel, ok := f.Uint()
		if !ok {
			return f.WrapperErr(fmt.Errorf("the param %s data type is not uint", f.CurParam))
		}

		val.SetUint(rel)

	case reflect.Float32, reflect.Float64:
		rel, ok := f.Float()
		if !ok {
			return f.WrapperErr(fmt.Errorf("the param %s data type is not float", f.CurParam))
		}

		val.SetFloat(rel)

	case reflect.Bool:
		rel, ok := f.Bool()
		if !ok {
			return f.WrapperErr(fmt.Errorf("the param %s data type is not bool", f.CurParam))
		}

		val.SetBool(rel)

	case reflect.String:
		rel, ok := f.Str()
		if !ok {
			return f.WrapperErr(fmt.Errorf("the param %s data type is not string", f.CurParam))
		}

		val.SetString(rel)

	}

	return nil
}

func buildValidRules(f *Field) rule.Inherits {
	tagVal := f.CurParam
	var bindArr rule.Inherits

	// check whether there have some parent inherit rule by dive
	if len(f.inherit) > 0 {
		bindArr = append(bindArr, f.inherit...)
	}

	if tagVal != "" && tagVal != "-" {
		// explode bing valid keys
		bindKeys := f.CurTag.Get(rule.KeyBind)
		if len(bindKeys) != 0 {
			rules := strings.Split(bindKeys, ",")
			for _, r := range rules {
				if strings.Contains(r, "=") {
					ruleArr := strings.Split(r, "=")
					bindArr = append(bindArr, rule.Info{
						Name: ruleArr[0],
						Val:  ruleArr[1],
					})
				} else {
					bindArr = append(bindArr, rule.Info{
						Name: r,
					})
				}
			}
		}
	}

	return bindArr
}

func valid(f *Field) (bool, error) {
	bindArr := buildValidRules(f)
	if len(bindArr) == 0 { // no parent rules、self no binding tags, will continue
		return false, nil
	}

	for i, r := range bindArr {
		f.CurRule = r

		if f.CurRule.Name == "omitempty" {
			if !f.Exist() { //
				return false, nil
			}

			// skip to next
			continue
		}

		// the data must map or slice when current valid is dive validator
		if f.CurRule.Name == "dive" && (f.selfType.Kind() == reflect.Map || f.selfType.Kind() == reflect.Slice) {
			tmpArr := bindArr[i+1:]
			var newInherit rule.Inherits
			for _, elem := range tmpArr {
				_, ok := rule.ValidatorFunc[elem.Name]
				if ok { // add level trace index
					elem.Level = append(elem.Level, ">")
				}
				newInherit = append(newInherit, elem)
			}

			f.inherit = newInherit
			return true, nil
		}

		fn := rule.ValidatorFunc[f.CurRule.Name]
		if fn == nil {
			panic(fmt.Sprintf("the valid func %s not exist", f.CurRule.Name))
		}

		//run valid
		if !fn(f) {
			return false, f.Err()
		}
	}

	return false, nil
}
