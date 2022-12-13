package server

import (
	"errors"
	"log"
	"micro-go-http-tpl/app"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/ntt360/gin/core/gracehttp"
)

func metrics(addr string) {
	http.Handle(app.Config.Metrics.DefaultPath(), promhttp.Handler())
	if err := gracehttp.ListenAndServe(addr, nil); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			log.Println(err)
		} else {
			panic(err)
		}
	}
}
