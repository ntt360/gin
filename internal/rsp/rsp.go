package rsp

// JSONVal http response common data.
type JSONVal struct {
	Code int         `json:"errno"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

type JSVal func(val *JSONVal)
