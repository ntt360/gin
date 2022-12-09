package gin

import (
	"github.com/ntt360/gin/internal/valid/binding"
	"io"
)

func (c *Context) Valid(obj any) error {
	b := binding.Default(c.Request.Method, c.ContentType())
	return c.ShouldBindWith(obj, b)
}

// ValidJSON is a shortcut for c.ValidWith(obj, binding.JSON).
func (c *Context) ValidJSON(obj any) error {
	return c.ShouldBindWith(obj, binding.JSON)
}

// ValidQuery is a shortcut for c.ValidWith(obj, binding.Query).
func (c *Context) ValidQuery(obj any) error {
	return c.ShouldBindWith(obj, binding.Query)
}

// ValidHeader is a shortcut for c.ShouldBindWith(obj, binding.Header).
func (c *Context) ValidHeader(obj any) error {
	return c.ShouldBindWith(obj, binding.Header)
}

// ValidWith binds the passed struct pointer using the specified binding engine.
// See the binding package.
func (c *Context) ValidWith(obj any, b binding.Binding) error {
	return b.Bind(c.Request, obj)
}

// ValidBodyWith is similar with ValidWith, but it stores the request
// body into the context, and reuse when it is called again.
//
// NOTE: This method reads the body before binding. So you should use
// ValidBodyWith for better performance if you need to call only once.
func (c *Context) ValidBodyWith(obj any, bb binding.BindingBody) (err error) {
	var body []byte
	if cb, ok := c.Get(BodyBytesKey); ok {
		if cbb, ok := cb.([]byte); ok {
			body = cbb
		}
	}
	if body == nil {
		body, err = io.ReadAll(c.Request.Body)
		if err != nil {
			return err
		}
		c.Set(BodyBytesKey, body)
	}
	return bb.BindBody(body, obj)
}
