package server

import (
	"crypto/tls"
	"errors"
	"github.com/ntt360/gin/core/gracehttp"
	"log"
	"net/http"
	"time"
)

func (s *Server) httpsServer() {
	server := gracehttp.NewServer(
		s.config.HTTPS.Listen,
		s.https.Engine(),
		time.Second*10,
		time.Second*10,
	)
	server.TLSConfig = &tls.Config{
		PreferServerCipherSuites: true,
		NextProtos:               []string{"h2", "http/1.1"},
	}
	err := server.ListenAndServeTLSOcsp(0, s.config.GetHTTPSCertFile(), s.config.GetHTTPSKeyFile())
	if err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			log.Println(err)
		} else {
			panic(err)
		}
	}
}
