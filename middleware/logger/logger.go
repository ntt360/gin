package logger

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/ntt360/gin/core/config"
	"github.com/ntt360/gin/core/gvalue"
	"micro-go-http-tpl/app"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"path"
	"runtime"
	"strings"
	"time"

	"github.com/ntt360/gin"

	"github.com/ntt360/gin/utils/uniqid"
)

// Logger app web server log instance.
// joinLines: 是否设置为上报数据到qbus队列
func Logger(config *config.Model) gin.HandlerFunc {
	conf := GinxLoggerrConfig{
		Formatter: defaultLogFormatter,
		SkipPaths: config.WebServerLog.SkipPaths,
	}

	err := initLogWriter(config)
	if err != nil {
		panic(err)
	}

	formatter := conf.Formatter
	notLogged := conf.SkipPaths

	var skip map[string]bool

	if length := len(notLogged); length > 0 {
		skip = make(map[string]bool, length)
		for _, pathItem := range notLogged {
			skip[pathItem] = true
		}
	}

	return func(c *gin.Context) {
		// Start timer
		start := time.Now()
		curPath := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		prefix := config.WebServerLog.TracePrefix
		if len(prefix) > 0 {
			prefix += "-"
		}

		// 自定义log TraceID, 添加时间戳方便查询
		logID := c.Request.Header.Get(gvalue.HttpHeaderLogIDKey)
		if logID == "" {
			logID = uniqid.LogID(prefix)
			c.Request.Header.Set(gvalue.HttpHeaderLogIDKey, logID)
		}

		// 添加响应头 Log Id
		c.Header("Ginx-Id", logID)
		// 程序崩溃错误收集
		defer panicCatch(c, config, start, raw, curPath, formatter)

		// Process request
		c.Next()

		// WebServerLog only when curPath is not being skipped
		if _, ok := skip[curPath]; !ok {
			errMsgs := c.Errors.ByType(gin.ErrorTypePrivate)
			outputErrMsg := "-"
			if len(errMsgs) > 0 {
				errMsg, err := errMsgs.MarshalJSON()
				if err != nil {
					return
				}
				outputErrMsg = string(errMsg)
			}

			writeData(c, config, start, raw, curPath, formatter, outputErrMsg)
		}
	}
}

func getLogDir(config *config.Model) string {
	logDir := strings.TrimRight(config.WebServerLog.Dir, "/")
	if !path.IsAbs(logDir) {
		logDir = strings.TrimRight(config.HomeDir, "/") + "/" + logDir
	}

	return logDir
}

func getLogBaseFileName(config *config.Model) string {
	return getLogDir(config) + "/" + strings.TrimLeft(config.WebServerLog.Name, "/")
}

// 加载Log配置
func loadLogWriter(config *config.Model) error {
	appLogName := getLogBaseFileName(config)
	appLog, err := os.OpenFile(appLogName, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	app.Locker.Lock()
	logOut = appLog
	app.Locker.Unlock()

	return nil
}

func initLogWriter(config *config.Model) error {
	switch config.WebServerLog.Output {
	case "file":
		err := loadLogWriter(config)
		if err != nil {
			return err
		}
	case "stdout":
		logOut = os.Stdout
	default:
		logOut = os.Stderr
	}

	return nil
}

func panicCatch(c *gin.Context, config *config.Model, start time.Time, raw string, path string, formatter LogFormatter) {
	if err := recover(); err != nil {
		var brokenPipe bool
		if ne, ok := err.(*net.OpError); ok {
			var se *os.SyscallError
			if errors.As(ne.Err, &se) {
				if strings.Contains(strings.ToLower(se.Error()), "broken pipe") || strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
					brokenPipe = true
				}
			}
		}

		stack := stack(3)
		httpRequest, _ := httputil.DumpRequest(c.Request, false)
		headers := strings.Split(string(httpRequest), "\r\n")
		for idx, header := range headers {
			current := strings.Split(header, ":")
			if current[0] == "Authorization" {
				headers[idx] = current[0] + ": *"
			}
		}

		panicErrString := ""
		if brokenPipe {
			panicErrString = fmt.Sprintf("%s\n%s", err, string(httpRequest))
		} else {
			panicErrString = fmt.Sprintf("panic recovered:\n%s\n%s\n%s", strings.Join(headers, "\r\n"), err, stack)
		}

		// 回写err, 供外层的中间件可能会使用到
		_ = c.Error(errors.New(panicErrString))

		// 上报数据 需要把数据变为单行数据，便于传输
		if config.WebServerLog.JoinLine {
			panicErrString = fmt.Sprintf("%#v", panicErrString)
		}

		c.AbortWithStatus(http.StatusInternalServerError)
		writeData(c, config, start, raw, path, formatter, panicErrString)
	}
}

func writeData(c *gin.Context, config *config.Model, start time.Time, raw string, path string, formatter LogFormatter, errMsg string) {
	param := gin.LogFormatterParams{
		Request: c.Request,
		Keys:    c.Keys,
		IsTerm:  config.WebServerLog.Output == "stdout" || config.WebServerLog.Output == "stderr",
	}

	// Stop timer
	param.TimeStamp = time.Now()
	param.Latency = param.TimeStamp.Sub(start)

	param.ClientIP = c.ClientIP()
	param.Method = c.Request.Method
	param.StatusCode = c.Writer.Status()

	param.ErrorMessage = errMsg

	param.BodySize = c.Writer.Size()

	if raw != "" {
		path = path + "?" + raw
	}

	param.Path = path

	paramWrapper := GinxLogFormatterParams{
		LogFormatterParams: param,
		Format:             config.WebServerLog.Format,
		Env:                config.Env,
		Idc:                config.IdcName,
		LocalIP:            config.LocalIP,
		Hostname:           config.Hostname,
		PrjName:            config.Name,
	}

	_, _ = fmt.Fprint(logOut, formatter(paramWrapper))

}

// stack returns a nicely formatted stack frame, skipping skip frames.
func stack(skip int) []byte {
	buf := new(bytes.Buffer) // the returned data
	// As we loop, we open files and read them. These variables record the currently
	// loaded file.
	var lines [][]byte
	var lastFile string
	for i := skip; ; i++ { // Skip the expected number of frames
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		// Print this much at least.  If we can't find the source, it won't show.
		_, _ = fmt.Fprintf(buf, "%s:%d (0x%x)\n", file, line, pc)
		if file != lastFile {
			data, err := os.ReadFile(file)
			if err != nil {
				continue
			}
			lines = bytes.Split(data, []byte{'\n'})
			lastFile = file
		}
		_, _ = fmt.Fprintf(buf, "\t%s: %s\n", function(pc), source(lines, line))
	}
	return buf.Bytes()
}

// source returns a space-trimmed slice of the n'th line.
func source(lines [][]byte, n int) []byte {
	n-- // in stack trace, lines are 1-indexed but our array is 0-indexed
	if n < 0 || n >= len(lines) {
		return dunno
	}
	return bytes.TrimSpace(lines[n])
}

// function returns, if possible, the name of the function containing the PC.
func function(pc uintptr) []byte {
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return dunno
	}
	name := []byte(fn.Name())
	// The name includes the path name to the package, which is unnecessary
	// since the file name is already included.  Plus, it has center dots.
	// That is, we see
	//	runtime/debug.*T·ptrmethod
	// and want
	//	*T.ptrmethod
	// Also the package path maybe contains dot (e.g. code.google.com/...),
	// so first eliminate the path prefix
	if lastSlash := bytes.LastIndex(name, slash); lastSlash >= 0 {
		name = name[lastSlash+1:]
	}
	if period := bytes.Index(name, dot); period >= 0 {
		name = name[period+1:]
	}
	name = bytes.Replace(name, centerDot, dot, -1)
	return name
}
