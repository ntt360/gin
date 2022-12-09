// Copyright 2014 Manu Martinez-Almeida. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package binding

import (
	"errors"
	"github.com/ntt360/gin/internal/valid/form"
	"github.com/ntt360/gin/valid"
	"net/http"
)

const defaultMemory = 32 << 20

type formBinding struct{}
type formMultipartBinding struct{}

var defGlobalErr = &valid.Error{
	Code: valid.CodeParamsErr,
	Type: valid.ErrorTypeGlobal,
	Msg:  "request data not valid",
}

func (formBinding) Name() string {
	return "form"
}

func (f formBinding) Bind(req *http.Request, obj any) error {
	if err := req.ParseForm(); err != nil {
		defGlobalErr.CauseErr = err
		return defGlobalErr
	}

	if err := req.ParseMultipartForm(defaultMemory); err != nil && !errors.Is(err, http.ErrNotMultipart) {
		defGlobalErr.CauseErr = err
		return defGlobalErr
	}

	data := make(map[string]any)
	for k, v := range req.Form {
		data[k] = v
	}

	return form.Validate(data, obj, f.Name())
}

func (formMultipartBinding) Name() string {
	return "form"
}

func (f formMultipartBinding) Bind(req *http.Request, obj any) error {
	if err := req.ParseMultipartForm(defaultMemory); err != nil {
		defGlobalErr.CauseErr = err
		return defGlobalErr
	}

	// merge multiform data and files data
	data := make(map[string]any)
	multiData := req.MultipartForm.Value
	for key, val := range multiData {
		data[key] = val
	}

	multiFiles := req.MultipartForm.File
	for key, val := range multiFiles {
		data[key] = val
	}

	return form.Validate(data, obj, f.Name())
}
