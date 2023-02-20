package server

import (
	"github.com/ntt360/gin/core/grpc/interceptor"
	"github.com/opentracing/opentracing-go"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
)

type GrpcRunner interface {
	NewGRPC(grpcServer *grpc.Server) *grpc.Server

	// Interceptors return user custom multi interceptor
	Interceptors() []grpc.UnaryServerInterceptor
}

func (s *Server) grpcServer() {
	lis, err := net.Listen("tcp", s.config.Grpc.Listen)
	if err != nil {
		panic(err)
	}
	var grpcServerOption []grpc.ServerOption

	var serverInterceptor []grpc.UnaryServerInterceptor

	// trace interceptor
	if s.config.Trace.Enable && opentracing.IsGlobalTracerRegistered() && s.config.Grpc.Server.Trace {
		tracer := opentracing.GlobalTracer()
		var opts []interceptor.Option

		if s.config.Trace.Rpc.LogReqParams {
			opts = append(opts, interceptor.LogReqParams())
		}

		if s.config.Trace.Rpc.LogRspPayload {
			opts = append(opts, interceptor.LogRspPayload())
		}

		serverInterceptor = append(serverInterceptor, interceptor.OpenTracingServerInterceptor(tracer, opts...))
	}

	// merge user custom interceptor
	if len(s.grpc.Interceptors()) > 0 {
		serverInterceptor = append(serverInterceptor, s.grpc.Interceptors()...)
	}

	grpcServerOption = append(
		grpcServerOption,
		grpc.ChainUnaryInterceptor(serverInterceptor...),
	)

	grpcS := grpc.NewServer(grpcServerOption...)
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig,
			syscall.SIGTERM,
			syscall.SIGINT)
		sNum := <-sig

		log.Println("grpc signal stop:", sNum)

		// TODO grpc 平滑重启感觉也需要加入graceful中，不然也有问题。目前没有使用grpc,暂时没有处理
		grpcS.GracefulStop()
	}()

	s.grpc.NewGRPC(grpcS)

	err = grpcS.Serve(lis)
	if err != nil {
		panic(err)
	}
}
