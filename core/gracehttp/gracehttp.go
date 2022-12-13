package gracehttp

import (
	"net/http"
	"time"
)

const (
	DefaultReadTimeout  = 60 * time.Second
	DefaultWriteTimeout = DefaultReadTimeout

	EnvDebug = "debug"
	EnvTest  = "test"
	EnvStage = "stage"
	EnvProd  = "prod"
)

var env = EnvProd

func SetEnv(mode string) {
	switch mode {
	case EnvDebug:
	case EnvTest:
	case EnvStage:
	case EnvProd:
		env = mode
	default:
		env = EnvDebug
	}

	env = mode
}

// ListenAndServe http
func ListenAndServe(addr string, handler http.Handler) error {
	return NewServer(addr, handler, DefaultReadTimeout, DefaultWriteTimeout).ListenAndServe()
}

// ListenAndServeTLS https
func ListenAndServeTLS(addr string, certFile string, keyFile string, handler http.Handler) error {
	return NewServer(addr, handler, DefaultReadTimeout, DefaultWriteTimeout).ListenAndServeTLSOcsp(0, certFile, keyFile)
}
