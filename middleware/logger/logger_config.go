package logger

import (
	"io"
)

type GinxLoggerrConfig struct {
	Formatter LogFormatter
	Output    io.Writer
	SkipPaths []string
}
