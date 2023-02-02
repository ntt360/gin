package json

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/ntt360/gin/internal/valid/rule"
	"github.com/tidwall/gjson"
)

func Validate(jsonMap gjson.Result, m any, validTagName string) error {
	val := reflect.ValueOf(m)

	if val.Type().Kind() == reflect.Ptr {
		val = val.Elem()
	}

	vf := &Field{
		BaseFiled: &rule.BaseFiled{
			ValidTagName: validTagName,
		},
		data: jsonMap,
		root: val.Type(),
	}

	// recurse
	_, err := recursiveValid(val, vf)
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

	case reflect.Map:
		// check map key type
		keyTyp := val.Type().Key().Kind()
		if keyTyp != reflect.String { // map key must string
			panic(fmt.Sprintf("map %s key must string", val.Type().Name()))
		}

		deepValid, err = valid(field)
		tmpMap := reflect.MakeMap(val.Type())

		var needValidKey bool
		var keysRules rule.Inherits

		cueParent := field.parent
		curSelf := field.self
		curTag := field.CurTag

		if deepValid { // valid in-depth
			if field.inherit.Contains("required") && !field.Exist() {
				return false, field.WrapperErr(fmt.Errorf("the param %s data required but data not exist", field.CurParam))
			}
			keysRules, needValidKey = checkValidMap(val, field)
		}

		// foreach map
		valData := field.data.Get(field.CurDataIdx)
		for elemKey := range valData.Map() {
			if deepValid && needValidKey {
				err = mapKeyValid(elemKey, keysRules, Field{
					self:   reflect.TypeOf(elemKey),
					parent: val.Type(),
					BaseFiled: &rule.BaseFiled{
						CurParam: field.CurParam,
						CurKey:   field.CurKey,
						CurTag:   field.CurTag,
					},
				})

				if err != nil {
					return false, err
				}
			}

			field.parent = val.Type()
			field.PushTrack(elemKey)
			field.CurTag = ""
			field.disableValid = deepValid

			// valid map value
			elemT := val.Type().Elem()
			field.self = elemT

			v := reflect.New(elemT)
			unValid, rErr := recursiveValid(v, field)
			if rErr != nil {
				return unValid, err
			}

			// try set map element value
			tmpMap.SetMapIndex(reflect.ValueOf(elemKey), reflect.Indirect(v))
			field.PopTrack()
		}

		// override the val data
		val.Set(tmpMap)

		// reset
		field.disableValid = false
		field.parent = cueParent
		field.self = curSelf
		field.CurTag = curTag

	case reflect.Slice, reflect.Array:

		// valid array self
		deepValid, err = valid(field)
		if err != nil {
			return false, err
		}

		if field.Exist() && !field.IsArray() {
			return false, field.WrapperErr(fmt.Errorf(" the param %s filed want array but data type not array", field.CurParam))
		}

		l := field.Len()
		tmpArr := reflect.MakeSlice(val.Type(), 0, l)

		// foreach array elements
		for j := 0; j < field.Len(); j++ {
			childTyp := val.Type().Elem()
			el := reflect.New(childTyp)

			field.PushTrack(strconv.Itoa(j))
			field.parent = val.Type()
			field.self = childTyp
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
		field.self = val.Type()
		field.CurKey = val.Type().Name()

		// valid struct self
		_, err = valid(field)
		if err != nil {
			return false, err
		}

		// the time.Time struct not deep valid
		if val.Type().ConvertibleTo(rule.TimeType) {
			tStr, ok := field.Str()
			if !ok {
				return true, nil
			}

			tVal, e := rule.FieldTimeVal(tStr, field.CurTag)
			if e != nil {
				return false, field.WrapperErr(fmt.Errorf("the params %s data can not covert to time", field.CurParam))
			}

			val.Set(reflect.ValueOf(tVal))
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
				field.self = vt.Type
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
			if err != nil {
				return unValid, rErr
			}

			field.PopTrack()
		}

	case reflect.Ptr: // give the pointer real elem
		if !val.Elem().IsValid() { // val child maybe is zero value or pointer value, such as *int
			v := reflect.New(val.Type().Elem())
			field.self = v.Type()
			unValid, rErr := recursiveValid(v.Elem(), field)
			if rErr != nil {
				return unValid, err
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
		valData := field.data.Get(field.CurDataIndex())
		if !valData.Exists() {
			return false, nil
		}

		val.Set(reflect.ValueOf(valData.Value()))
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

// 判断是否需要验证 map 数据类型的 key
func checkValidMap(val reflect.Value, field *Field) (rule.Inherits, bool) {
	needValidKey := field.inherit.Contains("keys")
	if (field.inherit.Contains("keys") && !field.inherit.Contains("endkeys")) || (!field.inherit.Contains("keys") && field.inherit.Contains("endkeys")) {
		panic(fmt.Sprintf("the %s missing keys or endkeys", val.Type().Name()))
	}

	var keysRules rule.Inherits
	if needValidKey {
		keysRules = field.inherit[field.inherit.IndexOf("keys")+1 : field.inherit.IndexOf("endkeys")]
		if len(keysRules) <= 0 {
			panic(fmt.Sprintf("the %s valid rules is empty between keys and endkeys", val.Type().Name()))
		}
	}

	// 验证 map key时候会rewrite、所以此处要回写
	if needValidKey {
		field.inherit = field.inherit[field.inherit.IndexOf("endkeys")+1:]
	}

	return keysRules, needValidKey
}

func mapKeyValid(key string, rules rule.Inherits, f Field) error {
	f.data = gjson.Parse(fmt.Sprintf("{\"key\": \"%s\"}", key))
	f.Track = []string{"key"}

	for _, r := range rules {
		f.CurRule = r
		if f.CurRule.Name == "omitempty" {
			if !f.Exist() {
				return nil
			}
		}

		// dive not allow between keys and endkeys
		if f.CurRule.Name == "dive" {
			panic(fmt.Sprintf("dive can not allow between keys and endkeys in param %s", f.CurKey))
		}

		fn := rule.ValidatorFunc[f.CurRule.Name]
		if fn == nil {
			panic(fmt.Sprintf("the valid func %s not exist", f.CurRule.Name))
		}

		//run valid
		if !fn(&f) {
			return f.Err()
		}
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
		if f.CurRule.Name == "dive" && (f.self.Kind() == reflect.Map || f.self.Kind() == reflect.Slice) {
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
