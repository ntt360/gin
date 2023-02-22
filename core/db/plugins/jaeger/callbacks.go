package jaeger

import (
	"context"
	"errors"
	"fmt"
	"github.com/ntt360/gin/core/opentrace"
	"strings"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

const (
	jaegerTraceKey = "jaeger_trace_span"
	Ctx            = "jaeger:ctx"
)

type bundle struct {
	Begin time.Time
	Sp    opentracing.Span
}

func before(db *gorm.DB) {
	ctx, ok := db.Statement.InstanceGet(Ctx)
	var jctx context.Context
	if ok {
		jctx = ctx.(context.Context)
	} else {
		jctx = context.Background()
	}

	// ignore no parent span trace
	if opentracing.SpanFromContext(jctx) == nil {
		return
	}

	// check custom operator name
	operatorName := "mysql"
	k := jctx.Value(opentrace.KeyAction)
	if v, yes := k.(string); yes {
		operatorName = v
	}

	sp, _ := opentracing.StartSpanFromContext(jctx, fmt.Sprintf("%s", operatorName))
	ext.DBType.Set(sp, "mysql")

	db.InstanceSet(jaegerTraceKey, bundle{
		Begin: time.Now(),
		Sp:    sp,
	})
}

func after(db *gorm.DB) {
	i, ok := db.InstanceGet(jaegerTraceKey)
	if !ok {
		return
	}

	j, _ := db.Plugins["plugin:jaeger"].(*Plugin)

	b, ok := i.(bundle)
	if !ok {
		return
	}
	defer b.Sp.Finish()
	durationTime := time.Since(b.Begin).Milliseconds()

	// 重组dsn，移除密码
	d, ok := db.Config.Dialector.(*mysql.Dialector)
	if ok {
		dsnArr := strings.Split(d.DSN, "@")
		dsnUser := dsnArr[0]
		if strings.Contains(dsnUser, ":") {
			dsnUser = strings.Split(dsnUser, ":")[0]
		}
		newDSN := fmt.Sprintf("%s:******@tcp%s", dsnUser, dsnArr[1])
		ext.DBInstance.Set(b.Sp, newDSN)
	}

	// 每条sql语句都要写，比较占资源，目前只处理大于200ms的sql语句上报
	if (j.IsAll() || (j.IsSlow() && durationTime >= int64(j.conf.SlowQueryTime))) && db.Error == nil {
		ext.DBStatement.Set(b.Sp, db.Dialector.Explain(db.Statement.SQL.String(), db.Statement.Vars...))
	}

	if db.Error != nil && !errors.Is(db.Error, gorm.ErrRecordNotFound) {
		ext.LogError(b.Sp, db.Error)
	}
}
