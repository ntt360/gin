// Copyright 2022 Gin Core Team. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package binding

import (
	"github.com/ntt360/gin/internal/valid/form"
	"net/http"
)

type headerBinding struct{}

func (headerBinding) Name() string {
	return "header"
}

func (h headerBinding) Bind(req *http.Request, obj any) error {
	data := make(map[string]any)
	for k, v := range req.Header {
		data[k] = v
	}

	return form.Validate(data, obj, h.Name())
}
