package server

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/ntt360/gin/core/config"
	"github.com/ntt360/gin/core/gvalue"

	"github.com/logrusorgru/aurora"
)

const (
	ipAll       = "0.0.0.0"
	paddingWith = 16
)

type Server struct {
	config *config.Base
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
			s.ls = append(s.ls, fmt.Sprintf(" - [%s] %s\n", centerPad("Task Jobs", 16), aurora.Bold(aurora.Green("Running"))))

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
			local, remote := getAddr(s.config.Grpc.Listen, "tcp://", aurora.GreenFg)
			s.ls = append(s.ls, fmt.Sprintf(" - [%s] Local: %-24s Network: %s\n", centerPad("Grpc Server", paddingWith), local, remote))
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
			local, remote := getAddr(s.config.HTTP.Listen, "http://", aurora.CyanFg)
			s.ls = append(s.ls, fmt.Sprintf(" - [%s] Local: %-24s Network: %s\n", centerPad("HTTP  Server", paddingWith), local, remote))
			s.http = runner

			go func() {
				defer s.wg.Done()
				s.httpServer()
			}()
		}
	}

	runners = append(runners, r)
}

func getAddr(addr string, prefix string, color aurora.Color) (aurora.Value, aurora.Value) {
	var localIP, remoteIP string
	if !strings.Contains(addr, ipAll) {
		local := aurora.Bold(aurora.Yellow(prefix + addr))
		remote := aurora.Bold(aurora.Yellow(prefix + addr))

		return local, remote
	}

	localIP = strings.Replace(addr, ipAll, "127.0.0.1", -1)
	remoteIP = strings.Replace(addr, ipAll, gvalue.LocalIP(), -1)

	var local, remote aurora.Value
	if color == aurora.YellowFg {
		local = aurora.Bold(aurora.Yellow(prefix + localIP))
		remote = aurora.Bold(aurora.Yellow(prefix + remoteIP))
	} else if color == aurora.GreenFg {
		local = aurora.Bold(aurora.Green(prefix + localIP))
		remote = aurora.Bold(aurora.Green(prefix + remoteIP))
	} else if color == aurora.CyanFg {
		local = aurora.Bold(aurora.Cyan(prefix + localIP))
		remote = aurora.Bold(aurora.Cyan(prefix + remoteIP))
	} else {
		local = aurora.Bold(prefix + localIP)
		remote = aurora.Bold(prefix + remoteIP)
	}

	return local, remote
}

func RegisterHttpsRunner(runner HttpRunner) {
	r := func(s *Server) {
		if s.config.HTTPS.Enable {
			s.wg.Add(1)
			local, remote := getAddr(s.config.HTTPS.Listen, "https://", aurora.CyanFg)
			s.ls = append(s.ls, fmt.Sprintf(" - [%s] Local: %-24s Network: %s\n", centerPad("HTTPS Server", paddingWith), local, remote))
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
func Run(config *config.Base) {
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

	// metrics
	if s.config.Metrics.Enable {
		s.wg.Add(1)
		p := 6061
		if s.config.Metrics.Port > 0 {
			p = s.config.Metrics.Port
		}
		addr := fmt.Sprintf("0.0.0.0:%d", p)
		local, remote := getAddr(addr, "http://", aurora.WhiteFg)
		s.ls = append(s.ls, fmt.Sprintf(" - [ %s ] Local: %-24s Network: %s\n", centerPad("Metrics Server", 14), local, remote))
		go func() {
			defer s.wg.Done()
			metrics(s.config.Metrics.DefaultPath(), addr)
		}()
	}

	// server will hang up
	s.wg.Wait()
}

func centerPad(title string, w int) string {
	return fmt.Sprintf("%*s", -w, fmt.Sprintf("%*s", (w+len(title))/2, title))
}
