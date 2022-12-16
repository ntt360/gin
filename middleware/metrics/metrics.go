package metrics

import (
	"github.com/ntt360/gin"
	"github.com/ntt360/gin/core/config"
	"github.com/prometheus/client_golang/prometheus"
	"sync"

	"strconv"
)

var pOnce sync.Once
var prom *Prometheus

type Prometheus struct {
	requestTotal *prometheus.CounterVec
	defaultPath  string
}

func NewPrometheus(config *config.Model) *Prometheus {
	pOnce.Do(func() {
		requestTotal := prometheus.NewCounterVec(prometheus.CounterOpts{
			Name:        "request_total",
			ConstLabels: map[string]string{"env": config.Env},
		}, []string{
			"status", "method", "uri",
		})
		prometheus.MustRegister(requestTotal)

		prom = &Prometheus{requestTotal: requestTotal, defaultPath: config.Metrics.DefaultPath()}

	})

	return prom
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
