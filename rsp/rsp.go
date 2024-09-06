package rsp

// JSONVal http response common data.
type JSONVal struct {
	Code int         `json:"errno"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

type JSVal func(val *JSONVal)

const (
	CodeOK  = 0
	CodeErr = 1

	MsgSuccess = "Success"
	MsgFailed  = "Failed"
)

func WithData(data any) JSVal {
	return func(val *JSONVal) {
		val.Data = data
	}
}

func WithMsg(msg string) JSVal {
	return func(val *JSONVal) {
		val.Msg = msg
	}
}

func WithCode(code int) JSVal {
	return func(val *JSONVal) {
		val.Code = code
	}
}
