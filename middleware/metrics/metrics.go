package metrics

import (
	"github.com/ntt360/gin"
	"github.com/ntt360/gin/core/config"
	"github.com/prometheus/client_golang/prometheus"

	"strconv"
)

type Prometheus struct {
	requestTotal *prometheus.CounterVec
	defaultPath  string
}

func NewPrometheus(config *config.Model) *Prometheus {
	requestTotal := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name:        "request_total",
		ConstLabels: map[string]string{"env": config.Env},
	}, []string{
		"status", "method", "uri",
	})
	prometheus.MustRegister(requestTotal)

	return &Prometheus{requestTotal: requestTotal, defaultPath: config.Metrics.DefaultPath()}
}

func (p *Prometheus) HandleFunc() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Next()
		if ctx.Request.URL.Path == p.defaultPath {
			return
		}

		status := strconv.Itoa(ctx.Writer.Status())
		method := ctx.Request.Method
		uri := ctx.Request.URL.Path

		p.requestTotal.With(prometheus.Labels{
			"status": status,
			"method": method,
			"uri":    uri,
		}).Inc()
	}
}
