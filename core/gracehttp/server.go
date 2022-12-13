package gracehttp

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/crypto/ocsp"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"
)

// 当前OCSP查询内存缓存
var cache []byte
var locker sync.RWMutex
var signalWatch sync.Once
var serverList []*Server
var signalChan = make(chan os.Signal, 1)

const (
	GracefulEnvironKey    = "IS_GRACEFUL"
	GracefulEnvironString = GracefulEnvironKey
	GracefulListenerFd    = 3
	OcspDefaultExpire     = time.Minute * 10
)

// Server HTTP server that supported graceful shutdown or restart
type Server struct {
	*http.Server

	listener       net.Listener
	originListener net.Listener
	isGraceful     bool
	certFile       string
	certKey        string
	ocspExpire     time.Duration
	shutdownChan   chan bool
}

func NewServer(addr string, handler http.Handler, readTimeout, writeTimeout time.Duration) *Server {
	isGraceful := false
	if os.Getenv(GracefulEnvironKey) != "" {
		isGraceful = true
	}

	return &Server{
		Server: &http.Server{
			Addr:         addr,
			Handler:      handler,
			ReadTimeout:  readTimeout,
			WriteTimeout: writeTimeout,
		},
		shutdownChan: make(chan bool),
		isGraceful:   isGraceful,
	}
}

func (srv *Server) ListenAndServe() error {
	addr := srv.Addr
	if addr == "" {
		srv.Addr = ":http"
	}

	ln, err := srv.getNetListener()
	if err != nil {
		return err
	}

	srv.listener = ln
	srv.originListener = ln
	if env == EnvDebug {
		fmt.Printf("The Server Is Runing: http://%s \n", srv.Addr)
	}

	return srv.Serve()
}

func (srv *Server) initServer(certFile string, keyFile string) {
	addr := srv.Addr
	if addr == "" {
		srv.Addr = ":https"
	}

	srv.certFile = certFile
	srv.certKey = keyFile
}

func (srv *Server) initConfig() *tls.Config {
	config := &tls.Config{}
	if srv.TLSConfig != nil {
		config = srv.TLSConfig.Clone()
	}

	if config.NextProtos == nil {
		config.NextProtos = []string{"http/1.1"}
	}

	return config
}

func (srv *Server) ListenAndServeTLSOcsp(expire time.Duration, certFile, keyFile string) error {
	srv.initServer(certFile, keyFile)
	if expire > 0 {
		srv.ocspExpire = expire
	}

	config := srv.initConfig()
	configHasCert := len(config.Certificates) > 0 || config.GetCertificate != nil
	if !configHasCert {
		config.GetCertificate = srv.GetCertificateWithOcsp
		srv.asyncOcspCache()
	}

	ln, err := srv.getNetListener()
	if err != nil {
		return err
	}

	srv.listener = tls.NewListener(ln, config)
	srv.originListener = ln
	if env == EnvDebug {
		fmt.Printf("The Server Is Runing: https://%s \n", srv.Addr)
	}

	return srv.Serve()
}

func (srv *Server) GetCertificateWithOcsp(*tls.ClientHelloInfo) (*tls.Certificate, error) {
	cert, err := tls.LoadX509KeyPair(srv.certFile, srv.certKey)
	if err != nil {
		return nil, err
	}

	if cache != nil {
		locker.RLock()
		cert.OCSPStaple = cache
		locker.RUnlock()
	}

	return &cert, nil
}

func (srv *Server) ListenAndServeTLS(certFile, keyFile string) error {
	srv.initServer(certFile, keyFile)
	config := srv.initConfig()

	configHasCert := len(config.Certificates) > 0 || config.GetCertificate != nil
	if !configHasCert || certFile != "" || keyFile != "" {
		config.Certificates = make([]tls.Certificate, 1)

		var err error
		config.Certificates[0], err = tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			return err
		}
	}

	ln, err := srv.getNetListener()
	if err != nil {
		return err
	}

	srv.listener = tls.NewListener(ln, config)
	srv.originListener = ln
	if env == EnvDebug {
		fmt.Printf("The Server Is Runing: https://%s \n", srv.Addr)
	}

	return srv.Serve()
}

func (srv *Server) Serve() error {
	// 插入serve列表
	serverList = append(serverList, srv)

	go signalWatch.Do(func() {
		handleSignals()
	})
	err := srv.Server.Serve(srv.listener)

	srv.logf("%s waiting for connections closed.", srv.Addr)
	<-srv.shutdownChan
	srv.logf("%s all connections closed.", srv.Addr)

	return err
}

func (srv *Server) getNetListener() (net.Listener, error) {
	var ln net.Listener
	var err error

	if srv.isGraceful {
		var rel map[string]int
		err = json.Unmarshal([]byte(os.Getenv(GracefulEnvironString)), &rel)
		if err != nil {
			return nil, fmt.Errorf("json decode error")
		}

		index, ok := rel[srv.Addr]
		if !ok {
			return nil, fmt.Errorf("%s fd index not found error", srv.Addr)
		}
		file := os.NewFile(uintptr(GracefulListenerFd+index), "")
		ln, err = net.FileListener(file)
		if err != nil {
			err = fmt.Errorf("net.FileListener error: %w", err)
			return nil, err
		}
		log.Printf("restart graceful server: %s", srv.Addr)
	} else {
		ln, err = net.Listen("tcp", srv.Addr)
		if err != nil {
			err = fmt.Errorf("net.Listen error: %w", err)
			return nil, err
		}
	}
	return ln, nil
}

func handleSignals() {
	var sig os.Signal

	signal.Notify(
		signalChan,
		syscall.SIGTERM,
		syscall.SIGINT,
		syscall.SIGUSR2,
	)

	for {
		sig = <-signalChan
		switch sig {
		case syscall.SIGTERM, syscall.SIGINT:
			log.Printf("received %s, graceful shutting down HTTP server.", sig.String())
			for _, srv := range serverList {
				srv.shutdownHTTPServer()
			}
		case syscall.SIGUSR2:
			log.Printf("received SIGUSR2, graceful restarting HTTP server.")

			pid, err := startNewProcess()
			if err != nil {
				log.Printf("start new process failed: %v, continue serving.", err)
				return
			}
			log.Printf("start new process successed, the new pid is %d.", pid)
		default:
		}
	}
}

func (srv *Server) shutdownHTTPServer() {
	if err := srv.Shutdown(context.Background()); err != nil {
		log.Printf("%s server shutdown error: %v", srv.Addr, err)
	} else {
		log.Printf("%s server shutdown success.", srv.Addr)
	}

	srv.shutdownChan <- true
}

// start new process to handle HTTP Connection
func startNewProcess() (uintptr, error) {
	// set graceful restart env flag
	var envs []string
	for _, value := range os.Environ() {
		if value != GracefulEnvironString {
			envs = append(envs, value)
		}
	}

	execSpec := &syscall.ProcAttr{
		Env:   envs,
		Files: []uintptr{os.Stdin.Fd(), os.Stdout.Fd(), os.Stderr.Fd()},
	}

	var fdIndex = make(map[string]int)
	for i, srv := range serverList {
		listenerFd, err := srv.getTCPListenerFd()
		if err != nil {
			return 0, fmt.Errorf("failed to get socket file descriptor: %w", err)
		}
		execSpec.Files = append(execSpec.Files, listenerFd)
		fdIndex[srv.Addr] = i
	}

	fdIndexData, _ := json.Marshal(fdIndex)
	execSpec.Env = append(execSpec.Env, fmt.Sprintf("%s=%s", GracefulEnvironKey, fdIndexData))
	fork, err := syscall.ForkExec(os.Args[0], os.Args, execSpec)
	if err != nil {
		return 0, fmt.Errorf("failed to forkexec: %w", err)
	}

	// 关掉父进程原先的server
	for _, srv := range serverList {
		srv.shutdownHTTPServer()
	}

	// 返回新进程id
	return uintptr(fork), nil
}

func (srv *Server) getTCPListenerFd() (uintptr, error) {
	file, err := srv.originListener.(*net.TCPListener).File()
	if err != nil {
		return 0, err
	}
	return file.Fd(), nil
}

func (srv *Server) logf(format string, args ...interface{}) {
	pids := strconv.Itoa(os.Getpid())
	format = "[pid " + pids + "] " + format

	if srv.ErrorLog != nil {
		srv.ErrorLog.Printf(format, args...)
	} else {
		log.Printf(format, args...)
	}
}

func (srv *Server) requestOCSP() error {
	cert, err := tls.LoadX509KeyPair(srv.certFile, srv.certKey)
	if err != nil {
		return err
	}

	if len(cert.Certificate) <= 1 {
		return errors.New("the cert have no leaf")
	}

	// 获取leaf证书，第一个证书
	x509Cert, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		return err
	}

	ocspServer := x509Cert.OCSPServer[0]
	if ocspServer == "" {
		return errors.New("ocsp server is empty")
	}

	x509Issuer, err := x509.ParseCertificate(cert.Certificate[1])
	if err != nil {
		return err
	}

	ocspRequest, err := ocsp.CreateRequest(x509Cert, x509Issuer, nil)
	if err != nil {
		return err
	}

	ocspRequestReader := bytes.NewReader(ocspRequest)
	c := &http.Client{
		Timeout: time.Second * 60,
	}

	httpResponse, err := c.Post(ocspServer, "application/ocsp-request", ocspRequestReader)
	if err != nil {
		return err
	}
	defer httpResponse.Body.Close()

	if httpResponse.StatusCode != http.StatusOK {
		return fmt.Errorf("csp rsp code not 200: %s", httpResponse.Status)
	}

	ocspRsp, err := io.ReadAll(httpResponse.Body)
	if err != nil {
		return err
	}

	locker.Lock()
	cache = ocspRsp
	locker.Unlock()

	return nil
}

func (srv *Server) asyncOcspCache() {
	go srv.scheduleOcsp()
}

func (srv *Server) scheduleOcsp() {
	dur := OcspDefaultExpire
	if srv.ocspExpire > 0 {
		dur = srv.ocspExpire
	}

	// do at once right now
	go func() {
		err := srv.requestOCSP()
		if err != nil {
			log.Println(err)
		}
	}()

	t := time.NewTicker(dur)
	defer t.Stop()

	for {
		<-t.C
		_ = srv.requestOCSP()
	}
}
