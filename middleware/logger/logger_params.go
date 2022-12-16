package logger

import (
	"encoding/json"
	"fmt"
	"github.com/ntt360/gin/core/gvalue"
	"github.com/ntt360/gin/internal/tty"
	"io"
)

var (
	dunno     = []byte("???")
	centerDot = []byte("·")
	dot       = []byte(".")
	slash     = []byte("/")
)

var logOut io.Writer

// LogFormatter ginx log formatter
type LogFormatter func(params GinxLogFormatterParams) string

var defaultLogFormatter = func(param GinxLogFormatterParams) string {

	// 生成全局唯一的loggerId
	logID := param.Request.Header.Get(gvalue.HttpHeaderLogIDKey)

	// ua
	ua := param.Request.UserAgent()
	if len(ua) == 0 {
		ua = "-"
	}

	refer := param.Request.Referer()
	if len(refer) <= 0 {
		refer = "-"
	}

	q := param.Request.URL.RawQuery
	if q == "" {
		q = "-"
	}

	idc := "-"
	if len(param.Idc) > 0 {
		idc = param.Idc
	}

	switch param.Format {
	case "json":
		data := map[string]interface{}{
			"ip":          param.ClientIP,
			"time":        param.TimeStamp.Format(gvalue.TimeRFC3339Milli),
			"log_id":      logID,
			"method":      param.Method,
			"path":        param.Request.URL.Path,
			"query":       q,
			"status_code": param.StatusCode,
			"body_size":   param.BodySize,
			"ua":          ua,
			"latency":     param.Latency.Milliseconds(),
			"err_msg":     param.ErrorMessage,
			"referer":     refer,
			"env":         param.Env,
			"idc":         idc,
			"prj":         param.PrjName,
			"local_ip":    param.LocalIP,
			"hostname":    param.Hostname,
		}
		encodeData, _ := json.Marshal(data)
		return string(encodeData) + "\n"
	default:
		var statusColor, methodColor, resetColor, logIDColor, ipColor string
		if param.IsOutputColor() {
			statusColor = param.StatusCodeColor()
			methodColor = param.MethodColor()
			resetColor = param.ResetColor()
			logIDColor = tty.Green
			ipColor = tty.Magenta
		}

		return fmt.Sprintf("[%s%s%s]\t\"%s\"\t%s%s%s\t%s%s%s\t%s\t%s\t%s %d %s\t%d\t\"%s\"\t%d\t%s\t%s\t%s\t%s\t%s\t%s%s%s\t%s\n",
			ipColor, param.ClientIP, resetColor,
			param.TimeStamp.Format(gvalue.TimeRFC3339Milli),
			logIDColor, logID, resetColor,
			methodColor, param.Method, resetColor,
			param.Request.URL.Path,
			q,
			statusColor, param.StatusCode, resetColor,
			param.BodySize,
			ua,
			param.Latency.Milliseconds(),
			param.ErrorMessage,
			refer,
			param.Env,
			idc,
			param.PrjName,
			ipColor, param.LocalIP, resetColor,
			param.Hostname,
		)
	}
}
