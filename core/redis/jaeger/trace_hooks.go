package jaeger

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-redis/redis/extra/rediscmd/v8"
	"github.com/go-redis/redis/v8"
	"github.com/ntt360/gin"
	"github.com/ntt360/gin/core/opentrace"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

var _ redis.Hook = TraceHooks{}

type TraceHooks struct {
	Addr string
}

func (c TraceHooks) BeforeProcess(ctx context.Context, cmd redis.Cmder) (context.Context, error) {
	unWrapCtx := context.Background()
	switch v := ctx.(type) {
	case *gin.Context:
		gCtx, ok := v.Get("traceCtx")
		if ok {
			unWrapCtx = gCtx.(context.Context)
		}
	case context.Context:
		unWrapCtx = v
	}

	// ignore no parent span trace
	if opentracing.SpanFromContext(unWrapCtx) == nil {
		return ctx, nil
	}

	// check custom operator name
	operatorName := "redis"
	k := ctx.Value(opentrace.KeyAction)
	if v, ok := k.(string); ok {
		operatorName = v
	}
	sp, spCtx := opentracing.StartSpanFromContext(unWrapCtx, fmt.Sprintf("%s %s", operatorName, cmd.FullName()))

	ext.DBType.Set(sp, "redis")
	ext.DBStatement.Set(sp, rediscmd.CmdString(cmd))

	sp.SetTag("db.addr", c.Addr)

	return spCtx, nil
}

func (c TraceHooks) AfterProcess(ctx context.Context, cmd redis.Cmder) error {
	sp := opentracing.SpanFromContext(ctx)
	if sp == nil {
		return nil
	}
	defer sp.Finish()

	// exclude redis nil err
	if cmd.Err() != nil && !errors.Is(cmd.Err(), redis.Nil) {
		ext.LogError(sp, cmd.Err())
	}

	return nil
}

func (c TraceHooks) BeforeProcessPipeline(ctx context.Context, cmds []redis.Cmder) (context.Context, error) {
	unWrapCtx := context.Background()
	switch v := ctx.(type) {
	case *gin.Context:
		gCtx, ok := v.Get("traceCtx")
		if ok {
			unWrapCtx = gCtx.(context.Context)
		}
	case context.Context:
		unWrapCtx = v
	}

	// ignore no parent span trace
	if opentracing.SpanFromContext(unWrapCtx) == nil {
		return ctx, nil
	}

	summary, cmdsString := rediscmd.CmdsString(cmds)

	sp, spCtx := opentracing.StartSpanFromContext(ctx, fmt.Sprintf("redis: %s", summary))
	ext.DBType.Set(sp, "redis")
	ext.DBStatement.Set(sp, cmdsString)
	sp.SetTag("db.addr", c.Addr)

	return spCtx, nil
}

func (c TraceHooks) AfterProcessPipeline(ctx context.Context, cmds []redis.Cmder) error {
	sp := opentracing.SpanFromContext(ctx)
	if sp == nil {
		return nil
	}
	defer sp.Finish()

	if cmds[0].Err() != nil && !errors.Is(cmds[0].Err(), redis.Nil) {
		ext.LogError(sp, cmds[0].Err())
	}

	return nil
}
