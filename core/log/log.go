package log

import (
	"errors"
	"os"
	"path"
	"path/filepath"

	"github.com/ntt360/gin/core/config"
	"github.com/ntt360/gin/core/gvalue"
	"github.com/sirupsen/logrus"
)

// NewLog app web new log.
func NewLog(logName string, conf *config.Model) (*logrus.Entry, error) {
	logger := logrus.New()

	// 日志输出级别
	initLogLevel(conf, logger)

	// 日志输出格式
	initLogFormat(conf, logger)

	// 设置日志输出
	err := initLogOutput(logName, conf, logger)
	if err != nil {
		return nil, err
	}

	return logger.WithFields(logrus.Fields{
		"module":   logName,
		"env":      conf.Env,
		"local_ip": conf.LocalIP,
		"hostname": conf.Hostname,
		"idc":      conf.IdcName,
		"prj":      conf.Name,
	}), nil
}

func initLogLevel(conf *config.Model, logger *logrus.Logger) {
	var level logrus.Level
	switch conf.Log.Level {
	case "debug":
		level = logrus.DebugLevel
	case "info":
		level = logrus.InfoLevel
	case "warn":
		level = logrus.WarnLevel
	case "error":
		level = logrus.ErrorLevel

	default:
		level = logrus.InfoLevel
	}

	logger.SetLevel(level)
}

func initLogFormat(conf *config.Model, logger *logrus.Logger) {
	switch conf.Log.Format {
	case "json":
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: gvalue.TimeRFC3339Milli,
		})
	default:
		txtFmt := &logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: gvalue.TimeRFC3339Milli,
		}
		if conf.Env != "dev" {
			txtFmt.DisableColors = true
		}
		logger.SetFormatter(txtFmt)
	}
}

func getLogFile(logName string, conf *config.Model) string {
	logPath := ""
	if path.IsAbs(logName) {
		logPath = logName
	} else {
		logPath = conf.HomeDir + "/logs/" + logName
	}

	return logPath
}

func initLogOutput(logName string, conf *config.Model, logger *logrus.Logger) error {
	// if log file name no ext, add default one
	logExt := filepath.Ext(logName)
	if len(logExt) <= 0 {
		logName = logName + ".log"
	}

	switch conf.Log.Output {
	case "stderr":
		logger.SetOutput(os.Stderr)
	case "file":
		if len(logName) <= 0 {
			return errors.New("log file output must config file name")
		}

		logF, err := initFileWriter(logName, conf)
		if err != nil {
			return err
		}
		logger.SetOutput(logF)
	default:
		logger.SetOutput(os.Stdout)
	}

	return nil
}

func initFileWriter(logName string, conf *config.Model) (*os.File, error) {
	logF, err := os.OpenFile(getLogFile(logName, conf), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	return logF, nil
}
