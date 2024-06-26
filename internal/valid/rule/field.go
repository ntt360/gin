package rule

import (
	"reflect"
)

const (
	KeyBind  = "binding"
	KeyMsg   = "msg"
	KeyRegex = "pattern"
)

type Field interface {
	// Root root element
	Root() reflect.Type

	// Parent  parent element
	Parent() reflect.Type

	// Self  wanted type
	Self() reflect.Type

	// Key Struct Key Name, self is struct key or inherit parent key name
	Key() string

	// StructField find relative another field key
	StructField(valType reflect.Type, namespace string) (Field, bool)

	// RuleName 当前验证的规则名称
	RuleName() string

	// RuleVal 当前验证规则值，可能不存在。
	RuleVal() string

	// CurDataIndex current data index
	CurDataIndex() string

	// SetDataIdx update data cur idx
	SetDataIdx(idx string)

	// Tag get current field CurTag
	Tag() reflect.StructTag

	// InheritTag inherit CurTag when dive valid
	InheritTag() reflect.StructTag

	// DefaultValExits 默认值是否存在
	DefaultValExits() bool

	// DefaultVal 默认值
	DefaultVal() string

	// Exist data exist or not
	Exist() bool

	// Empty data is empty or not
	Empty() bool
	
	// IsBool data is bool value or not
	IsBool() bool

	// IsArray data is array type or not, not support default value
	IsArray() bool

	// IsObject data is map or struct
	IsObject() bool

	// IsNumeric data is numeric
	IsNumeric() bool

	// Str try to get str data
	Str() (string, bool)

	// MustStr return str ignore err
	MustStr() string

	// Int try to get int data
	Int() (int64, bool)

	// MustInt return int val will ignore err or not exist
	MustInt() int64

	// Uint try to get uint data
	Uint() (uint64, bool)

	// MustUint return the confirmed val
	MustUint() uint64

	// Float try to get float data
	Float() (float64, bool)

	// MustFloat return float val will ignore the val not exist
	MustFloat() float64

	// Len try to get data len
	Len() int

	// Bool return data bool value
	Bool() (bool, bool)

	// MustBool return bool value ignore the error or the val not exist
	MustBool() bool

	// Raw get the data raw string
	Raw() string

	// ParamKey the value just is one of: json, form, header
	ParamKey() string

	// MapKeys json may keys array, them json map key only support string
	MapKeys() []string

	// ToString try to convert the value type to string
	ToString() string
}
