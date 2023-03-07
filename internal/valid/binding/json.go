// Copyright 2014 Manu Martinez-Almeida. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package binding

import (
	"bytes"
	j "encoding/json"
	"github.com/ntt360/gin/internal/valid/json"
	"github.com/tidwall/gjson"
	"io"
	"net/http"
)

type jsonBinding struct{}

func (j jsonBinding) Name() string {
	return "json"
}

func (j jsonBinding) Bind(req *http.Request, obj any) error {
	if req == nil || req.Body == nil {
		defGlobalErr.Msg = "invalid request"

		return defGlobalErr
	}

	bodyData, _ := io.ReadAll(req.Body)

	// rewrite data
	req.Body = io.NopCloser(bytes.NewBuffer(bodyData))

	return decodeJSON(bodyData, obj, j.Name())
}

func (j jsonBinding) BindBody(body []byte, obj any) error {
	return decodeJSON(body, obj, j.Name())
}

func decodeJSON(content []byte, obj any, validTypeName string) error {

	if len(content) > 0 && !j.Valid(content) {
		defGlobalErr.Msg = "request data is not valid json"

		return defGlobalErr
	}

	// both valid and bind data to obj
	return json.Validate(gjson.ParseBytes(content), obj, validTypeName)
}
