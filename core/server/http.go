package server

import (
	"errors"
	"github.com/ntt360/gin"
	"github.com/ntt360/gin/core/gracehttp"
	"log"
	"net/http"
)

type HttpRunner interface {
	Engine() *gin.Engine
}

func (s *Server) httpServer() {
	err := gracehttp.ListenAndServe(s.config.HTTP.Listen, s.http.Engine())
	if errors.Is(err, http.ErrServerClosed) {
		log.Println(err)
	} else {
		panic(err)
	}
}
