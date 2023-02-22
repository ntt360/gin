package tracer

import (
	"context"
	"fmt"
	"os"

	"github.com/ntt360/gin"
	"github.com/ntt360/gin/core/config"
	"github.com/ntt360/gin/core/gvalue"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/uber/jaeger-client-go"
)

func Inject(appConf *config.Base) gin.HandlerFunc {
	var filterPaths = make(map[string]struct{})
	for _, val := range appConf.Trace.SkipPaths {
		filterPaths[val] = struct{}{}
	}

	idc := os.Getenv("SYS_IDC_NAME")

	return func(ctx *gin.Context) {
		if _, ok := filterPaths[ctx.Request.URL.Path]; ok {
			ctx.Next()
			return
		}

		spCtx, e := opentracing.GlobalTracer().Extract(
			opentracing.HTTPHeaders,
			opentracing.HTTPHeadersCarrier(ctx.Request.Header),
		)

		var rootSP opentracing.Span
		if e != nil {
			rootSP = opentracing.StartSpan(fmt.Sprintf("http %s", ctx.Request.URL.Path))
		} else {
			rootSP = opentracing.StartSpan(fmt.Sprintf("http %s", ctx.Request.URL.Path), opentracing.ChildOf(spCtx))
		}
		defer rootSP.Finish()

		if sc, ok := rootSP.Context().(jaeger.SpanContext); ok {
			ctx.Request.Header.Set(gvalue.HttpHeaderLogIDKey, sc.TraceID().String())
		}

		rootSP.SetTag("service.env", appConf.Env)
		rootSP.SetTag("service.summary", appConf.Summary)
		rootSP.SetTag("service.idc", idc)

		ext.HTTPUrl.Set(rootSP, ctx.Request.URL.String())
		ext.HTTPMethod.Set(rootSP, ctx.Request.Method)
		ext.PeerHostname.Set(rootSP, ctx.Request.Host)
		ext.PeerAddress.Set(rootSP, ctx.ClientIP())

		// 向后传递
		ctx.Set("traceCtx", opentracing.ContextWithSpan(context.Background(), rootSP))
		ctx.Next()

		code := ctx.Writer.Status()
		ext.HTTPStatusCode.Set(rootSP, uint16(code))

		if len(ctx.Errors) > 0 {
			err := ctx.Errors[0]
			ext.LogError(rootSP, err)
		}
	}
}
