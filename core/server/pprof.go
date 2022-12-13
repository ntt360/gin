package server

import (
	"errors"
	"log"
	"net/http"
	_ "net/http/pprof"

	"github.com/ntt360/gin/core/gracehttp"
)

func (s *Server) pprof(localIP string) {
	err := gracehttp.ListenAndServe(localIP, nil)
	if err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			log.Println(err)
		} else {
			panic(err)
		}
	}
}
