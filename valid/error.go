package valid

import (
	"fmt"
	"io"
)

type ErrorType int

const (
	ErrorTypeGlobal ErrorType = 1
	ErrorTypeField  ErrorType = 2

	CodeParamsErr = 1 // param code number
)

// Error validator error support code number、filed param、filed key、
type Error struct {
	// error code num not used at present! all code return 1
	Code int

	// the front define name
	Param string

	// RuleName current error validator name
	RuleName string

	// the struct key name if error on struct field
	Key string

	// error msg if define custom msg will used preferred
	Msg string

	// Type the error type: global error not happened on the detail field , not contain any field info, the Key and param maybe empty
	Type ErrorType

	// CauseErr the real error msg used for log
	CauseErr error
}

func (e *Error) Error() string {
	return e.Msg
}

// Unwrap provides compatibility for Go 1.13 error chains.
func (e *Error) Unwrap() error { return e.CauseErr }

func (e *Error) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			var msg string
			if e.Type == ErrorTypeField {
				msg = fmt.Sprintf("the related param %s not valid in %s, the related Field Key is %s", e.Param, e.RuleName, e.Key)
				_, _ = fmt.Fprintf(s, "%s \n", msg)
			} else {
				_, _ = fmt.Fprintf(s, "%s \n", e.Msg)
			}

			if e.CauseErr != nil {
				_, _ = io.WriteString(s, e.CauseErr.Error())
			}

			return
		}
		fallthrough
	case 's', 'q':
		_, _ = io.WriteString(s, e.Error())
	}
}
