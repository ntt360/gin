package gin

import (
	"context"
	"io"
	"net/http"
	"net/url"

	"github.com/ntt360/gin/internal/rsp"
	"github.com/ntt360/gin/internal/valid/binding"
	r "github.com/ntt360/gin/rsp"
)

const TraceContextKey = "traceCtx"

var (
	jsonpCallbackRegex = "^[\\w-.]{1,64}$"
	jsonpCallbackName  = "callback"
)

func SetJsonpCallbackName(name string) {
	jsonpCallbackName = name
}

func SetJsonpCallbackRegex(regex string) {
	jsonpCallbackRegex = regex
}

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

// JSONOk response json success data.
func (c *Context) JSONOk(val ...rsp.JSVal) {
	rel := &rsp.JSONVal{
		Code: r.CodeOK,
		Msg:  r.MsgSuccess,
	}

	for _, jsonVal := range val {
		jsonVal(rel)
	}

	c.JSON(http.StatusOK, rel)
}

// JSONErr response json error data.
func (c *Context) JSONErr(val ...rsp.JSVal) {
	rel := &rsp.JSONVal{
		Code: r.CodeErr,
		Msg:  r.MsgFailed,
	}

	for _, jsonVal := range val {
		jsonVal(rel)
	}

	c.JSON(http.StatusOK, rel)
}

// JSONPOk response jsonp success data.
func (c *Context) JSONPOk(val ...rsp.JSVal) {
	rel := &rsp.JSONVal{
		Code: r.CodeOK,
		Msg:  r.MsgSuccess,
	}

	for _, jsonVal := range val {
		jsonVal(rel)
	}

	c.JSONP(http.StatusOK, rel)
}

// JSONPErr response jsonp error data.
func (c *Context) JSONPErr(val ...rsp.JSVal) {
	rel := &rsp.JSONVal{
		Code: r.CodeErr,
		Msg:  r.MsgFailed,
		Data: nil,
	}

	for _, jsonVal := range val {
		jsonVal(rel)
	}

	c.JSONP(http.StatusOK, rel)
}

// TraceCtx get trace context with key
func (c *Context) TraceCtx() context.Context {
	tCtx, ok := c.Get(TraceContextKey)
	if ok {
		return tCtx.(context.Context)
	}

	return context.TODO()
}

// QueryAll return all Get Query keys with values
func (c *Context) QueryAll() url.Values {
	c.initQueryCache()

	return c.queryCache
}
