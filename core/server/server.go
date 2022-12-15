package server

import (
	"fmt"
	"github.com/ntt360/gin/core/config"
	"micro-go-http-tpl/app"
	"strings"
	"sync"
	"time"

	"github.com/logrusorgru/aurora"
)

type Server struct {
	config *config.Model
	wg     sync.WaitGroup
	ls     []string

	https, http HttpRunner
	grpc        GrpcRunner
	task        TaskRunner
}

var runners []RunnerOption

type RunnerOption func(s *Server)

func RegisterTaskRunner(task TaskRunner) {
	r := func(s *Server) {
		// 注册task任务，需要配置 app.Config.yaml task[enable] = true
		if s.config.Task.Enable {
			s.wg.Add(1)
			s.ls = append(s.ls, fmt.Sprintf(" - Task Jobs: %s\n", aurora.Bold(aurora.Green("Running"))))
			go func() {
				task.Runner(&s.wg)
			}()
		}
	}

	runners = append(runners, r)
}

func RegisterGrpcRunner(runner GrpcRunner) {
	r := func(s *Server) {
		// grpc server
		if s.config.Grpc.Enable {
			s.wg.Add(1)
			s.ls = append(s.ls, fmt.Sprintf(" - Grpc  Server: %s\n", aurora.Bold(aurora.Cyan("tcp://"+s.config.Grpc.Listen))))
			s.grpc = runner

			go func() {
				defer s.wg.Done()
				s.grpcServer()
			}()
		}
	}

	runners = append(runners, r)
}

func RegisterHttpRunner(runner HttpRunner) {
	r := func(s *Server) {
		if s.config.HTTP.Enable {
			s.wg.Add(1)
			s.ls = append(s.ls, fmt.Sprintf(" - HTTP  Server: %s\n", aurora.Bold(aurora.Cyan("http://"+s.config.HTTP.Listen))))
			s.http = runner

			go func() {
				defer s.wg.Done()
				s.httpServer()
			}()
		}
	}

	runners = append(runners, r)
}

func RegisterHttpsRunner(runner HttpRunner) {
	r := func(s *Server) {
		if s.config.HTTPS.Enable {
			s.wg.Add(1)
			s.ls = append(s.ls, fmt.Sprintf(" - HTTPS Server: %s\n", aurora.Bold(aurora.Cyan("https://"+s.config.HTTPS.Listen))))
			s.https = runner

			go func() {
				defer s.wg.Done()
				s.httpsServer()
			}()
		}
	}

	runners = append(runners, r)
}

// Run init all server
func Run(config *config.Model) {
	s := &Server{
		config: config,
		wg:     sync.WaitGroup{},
	}

	for _, runner := range runners {
		runner(s)
	}

	go func() {
		// TODO 如此输出方式可以深入框架优化
		time.Sleep(time.Second * 1)
		fmt.Printf("\n\nApp Is Running At :\n%s\n", strings.Join(s.ls, ""))
	}()

	// pprof
	if s.config.PProf.Enable {
		s.wg.Add(1)
		p := 6060
		if s.config.PProf.Port > 0 {
			p = s.config.PProf.Port
		}

		addr := fmt.Sprintf("0.0.0.0:%d", p)
		s.ls = append(s.ls, fmt.Sprintf(" - PProf Server: %s\n", aurora.Bold(aurora.Cyan("http://"+addr))))

		go func() {
			defer s.wg.Done()
			s.pprof(addr)
		}()
	}

	// metrics
	if s.config.Metrics.Enable {
		s.wg.Add(1)
		p := 6061
		if s.config.Metrics.Port > 0 {
			p = app.Config.Metrics.Port
		}
		addr := fmt.Sprintf("0.0.0.0:%d", p)
		s.ls = append(s.ls, fmt.Sprintf(" - metrics Server: %s\n", aurora.Bold(aurora.Cyan("http://"+addr))))
		go func() {
			defer s.wg.Done()
			metrics(addr)
		}()
	}

	// server will hang up
	s.wg.Wait()
}
