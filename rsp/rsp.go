package rsp

import "github.com/ntt360/gin/internal/rsp"

const (
	CodeOK  = 0
	CodeErr = 1

	MsgSuccess = "Success"
	MsgFailed  = "Failed"
)

func WithData(data any) rsp.JSVal {
	return func(val *rsp.JSONVal) {
		val.Data = data
	}
}

func WithMsg(msg string) rsp.JSVal {
	return func(val *rsp.JSONVal) {
		val.Msg = msg
	}
}

func WithCode(code int) rsp.JSVal {
	return func(val *rsp.JSONVal) {
		val.Code = code
	}
}
