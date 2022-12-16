package e

import (
	"fmt"
	"io"
)

type CodeErr struct {
	code  int
	msg   string
	cause error
}

func (e *CodeErr) Error() string {
	return e.msg
}

func (e *CodeErr) Code() int {
	return e.code
}

func (e *CodeErr) Cause() error { return e.cause }

// Unwrap provides compatibility for Go 1.13 error chains.
func (e *CodeErr) Unwrap() error { return e.cause }

func (e *CodeErr) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			_, _ = fmt.Fprintf(s, "%+v \n", e.Cause())
			_, _ = io.WriteString(s, e.msg)
			return
		}
		fallthrough
	case 's', 'q':
		_, _ = io.WriteString(s, e.Error())
	}
}

func wrapErrf(err error, code int, format string, a ...any) error {
	e := &CodeErr{
		code:  code,
		msg:   fmt.Sprintf(format, a...),
		cause: err,
	}

	return &withStack{
		error: e,
		stack: callers(4),
	}
}

// baseErr 系统错误
// code: 错误码
// stackSkip 栈起始记录的位置
func baseErr(msg string, code int) error {
	e := &CodeErr{
		code: code,
		msg:  msg,
	}

	return &withStack{
		error: e,
		stack: callers(4),
	}
}
