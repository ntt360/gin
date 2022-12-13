package gormlogger

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"gorm.io/gorm/logger"
	"gorm.io/gorm/utils"

	"github.com/ntt360/gin/core/config"
	"github.com/ntt360/gin/core/gvalue"
)

var logr *logrus.Entry

func init() {
	logrus.SetFormatter(&logrus.JSONFormatter{TimestampFormat: time.RFC1123Z})
	//	logr = logrus.New()
	logr = logrus.WithField("module", "gorm")
}

var (
	traceLogLen = 6
)

func New(env string, writer logger.Writer, config logger.Config) logger.Interface {
	var (
		infoStr      = "%s %s %s\n[info] "
		warnStr      = "%s %s %s\n[warn] "
		errStr       = "%s %s %s\n[error] "
		traceStr     = "%s\n[%.3fms] [rows:%v] %s %s %s"
		traceWarnStr = "%s %s\n[%.3fms] [rows:%v] %s %s %s"
		traceErrStr  = "%s %s\n[%.3fms] [rows:%v] %s %s %s"
	)

	if config.Colorful {
		infoStr = logger.Green + "%s " + logger.Reset + "%s " + logger.Green + "%s\n" + logger.Reset + "[info] " + logger.Reset
		warnStr = logger.BlueBold + "%s " + logger.BlueBold + "%s " + logger.BlueBold + "%s\n" + logger.Magenta + "[warn] " + logger.Reset
		errStr = logger.Magenta + "%s " + logger.Magenta + "%s " + logger.Reset + "%s\n" + logger.Red + "[error] " + logger.Reset
		traceStr = logger.Green + "%s\n" + logger.Reset + logger.Yellow + "[%.3fms] " + logger.BlueBold + "[rows:%v]" + logger.Reset + " %s" + logger.Yellow + " %s" + logger.Blue + " %s" + logger.Reset
		traceWarnStr = logger.Green + "%s " + logger.Yellow + "%s\n" + logger.Reset + logger.RedBold + "[%.3fms] " + logger.Yellow + "[rows:%v]" + logger.Magenta + " %s" + logger.Yellow + " %s" + logger.Blue + " %s" + logger.Reset
		traceErrStr = logger.RedBold + "%s " + logger.MagentaBold + "%s\n" + logger.Reset + logger.Yellow + "[%.3fms] " + logger.BlueBold + "[rows:%v]" + logger.Reset + " %s" + logger.Yellow + " %s" + logger.Blue + " %s" + logger.Reset
	}

	return &gormlogger{
		Writer:       writer,
		Config:       config,
		infoStr:      infoStr,
		warnStr:      warnStr,
		errStr:       errStr,
		traceStr:     traceStr,
		traceWarnStr: traceWarnStr,
		traceErrStr:  traceErrStr,
		env:          env,
	}
}

type gormlogger struct {
	env string
	logger.Writer
	logger.Config
	infoStr, warnStr, errStr            string
	traceStr, traceErrStr, traceWarnStr string
}

// LogMode log mode
func (l *gormlogger) LogMode(level logger.LogLevel) logger.Interface {
	newlogger := *l
	newlogger.LogLevel = level
	return &newlogger
}

// Printf Print
func (l *gormlogger) Printf(s string, v ...interface{}) {
	if l.env == config.Dev {
		sl := log.New(os.Stdout, "\r\n", log.LstdFlags)
		_ = sl.Output(2, fmt.Sprintf(s, v...))
		return
	}
	logr = logrus.WithFields(
		logrus.Fields{
			"module":    "gorm",
			"type":      "sql",
			"file_line": v[0],
			"trace_id":  v[1],
			"ip":        v[2],
		},
	)
	if strings.Contains(s, "[info]") {
		logr.Info()
	} else if strings.Contains(s, "[warn]") {
		logr.Warn()
	} else if strings.Contains(s, "[error]") {
		logr.Error()
	} else {
		if len(v) == traceLogLen { // traceStr
			logr = logrus.WithFields(
				logrus.Fields{
					"module":       "gorm",
					"type":         "sql",
					"file_line":    v[0],
					"duration":     v[1],
					"row_affected": v[2],
					"sql":          v[3],
					"trace_id":     v[4],
					"ip":           v[5],
				},
			)
		} else if len(v) > traceLogLen {
			logr = logrus.WithFields(
				logrus.Fields{
					"module":        "gorm",
					"type":          "sql",
					"file_line":     v[0],
					"duration":      v[2],
					"rows_affected": v[3],
					"sql":           v[4],
					"note":          v[1],
					"trace_id":      v[5],
					"ip":            v[6],
				},
			)
		}
		logr.Info()
	}
}

// Info print info
func (l *gormlogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Info {
		var traceID, ip string
		if v := ctx.Value(gvalue.DBBaggageKey); v != nil {
			if c, ok := v.(gvalue.Baggage); ok {
				traceID = c.TraceID
				ip = c.IP
			}
		}
		l.Printf(l.infoStr+msg, append([]interface{}{utils.FileWithLineNum(), traceID, ip}, data...)...)
	}
}

// Warn print warn messages
func (l *gormlogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Warn {
		var traceID, ip string
		if v := ctx.Value(gvalue.DBBaggageKey); v != nil {
			if c, ok := v.(gvalue.Baggage); ok {
				traceID = c.TraceID
				ip = c.IP
			}
		}
		l.Printf(l.warnStr+msg, append([]interface{}{utils.FileWithLineNum(), traceID, ip}, data...)...)
	}
}

// Error print error messages
func (l *gormlogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Error {
		var traceID, ip string
		if v := ctx.Value(gvalue.DBBaggageKey); v != nil {
			if c, ok := v.(gvalue.Baggage); ok {
				traceID = c.TraceID
				ip = c.IP
			}
		}
		l.Printf(l.errStr+msg, append([]interface{}{utils.FileWithLineNum(), traceID, ip}, data...)...)
	}
}

// Trace print sql message
func (l *gormlogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.LogLevel > logger.Silent {
		var traceID, ip string
		if v := ctx.Value(gvalue.DBBaggageKey); v != nil {
			if c, ok := v.(gvalue.Baggage); ok {
				traceID = c.TraceID
				ip = c.IP
			}
		}
		elapsed := time.Since(begin)
		switch {
		case err != nil && l.LogLevel >= logger.Error:
			sql, rows := fc()
			if rows == -1 {
				l.Printf(l.traceErrStr, utils.FileWithLineNum(), err, float64(elapsed.Nanoseconds())/1e6, "-", sql, traceID, ip)
			} else {
				l.Printf(l.traceErrStr, utils.FileWithLineNum(), err, float64(elapsed.Nanoseconds())/1e6, rows, sql, traceID, ip)
			}
		case elapsed > l.SlowThreshold && l.SlowThreshold != 0 && l.LogLevel >= logger.Warn:
			sql, rows := fc()
			slowLog := fmt.Sprintf("SLOW SQL >= %v", l.SlowThreshold)
			if rows == -1 {
				l.Printf(l.traceWarnStr, utils.FileWithLineNum(), slowLog, float64(elapsed.Nanoseconds())/1e6, "-", sql, traceID, ip)
			} else {
				l.Printf(l.traceWarnStr, utils.FileWithLineNum(), slowLog, float64(elapsed.Nanoseconds())/1e6, rows, sql, traceID, ip)
			}
		case l.LogLevel == logger.Info:
			sql, rows := fc()
			if rows == -1 {
				l.Printf(l.traceStr, utils.FileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, "-", sql, traceID, ip)
			} else {
				l.Printf(l.traceStr, utils.FileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, rows, sql, traceID, ip)
			}
		}
	}
}
