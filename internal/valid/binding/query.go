// Copyright 2017 Manu Martinez-Almeida. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package binding

import (
	"github.com/ntt360/gin/internal/valid/form"
	"net/http"
)

type queryBinding struct{}

func (queryBinding) Name() string {
	return "query"
}

func (q queryBinding) Bind(req *http.Request, obj any) error {
	values := req.URL.Query()
	data := make(map[string]any)
	for k, v := range values {
		data[k] = v
	}

	// query use form valid tag
	return form.Validate(data, obj, q.Name())
}
