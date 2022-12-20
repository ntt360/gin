package server

import (
	"errors"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/ntt360/gin/core/gracehttp"
)

func metrics(defaultPath, addr string) {
	http.Handle(defaultPath, promhttp.Handler())
	if err := gracehttp.ListenAndServe(addr, nil); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			log.Println(err)
		} else {
			panic(err)
		}
	}
}
