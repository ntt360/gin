package e

import (
	"fmt"
)

// Common err code
const (
	CodeOk            = 0 // Success
	CodeServerErr     = 1 // 服务器错误
	CodeParamNotValid = 2 // 参数验证失败
)

const (
	MsgServerErr = "server err"
)

// WithStack 只提供默认栈封装，不包含其他处理逻辑
func WithStack(err error) error {
	return &withStack{
		error: err,
		stack: callers(),
	}
}

// NewCodeErrf 自定义Code码的错误消息
func NewCodeErrf(code int, format string, a ...any) error {
	return baseErr(fmt.Sprintf(format, a...), code)
}

func WrapCodeErrf(err error, code int, format string, a ...any) error {
	return wrapErrf(err, code, format, a...)
}

func WrapParamErrf(err error, format string, a ...any) error {
	return wrapErrf(err, CodeParamNotValid, format, a...)
}

func WrapSysErrf(err error, format string, a ...any) error {
	return wrapErrf(err, CodeServerErr, format, a...)
}

func WrapDefaultSysErr(err error) error {
	return wrapErrf(err, CodeServerErr, MsgServerErr)
}

// NewParamErrf 参数类型错误，自定义消息内容，支持格式化内容
func NewParamErrf(format string, a ...any) error {
	return baseErr(fmt.Sprintf(format, a...), CodeParamNotValid)
}

// DefaultSysErr 默认系统错误，即提供默认的错误码，和错误描述
func DefaultSysErr() error {
	return baseErr(MsgServerErr, CodeServerErr)
}

// NewSysErrf 系统类型错误，支持自定义错误格式
func NewSysErrf(format string, a ...any) error {
	return baseErr(fmt.Sprintf(format, a...), CodeServerErr)
}
