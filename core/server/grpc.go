package server

import (
	"github.com/ntt360/gin/core/grpc/interceptor"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
)

type GrpcRunner interface {
	NewGRPC(grpcServer *grpc.Server) *grpc.Server
}

func (s *Server) grpcServer() {
	// TODO grpc 平滑重启感觉也需要加入graceful中，不然也有问题。目前没有使用grpc,暂时没有处理
	lis, err := net.Listen("tcp", s.config.Grpc.Listen)
	if err != nil {
		panic(err)
	}
	grpcServerOption := []grpc.ServerOption{
		grpc.UnaryInterceptor(interceptor.DBContext),
	}
	grpcS := grpc.NewServer(grpcServerOption...)
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig,
			syscall.SIGTERM,
			syscall.SIGINT)
		s := <-sig
		log.Println("grpc signal stop:", s)
		grpcS.GracefulStop()
	}()

	s.grpc.NewGRPC(grpcS)

	err = grpcS.Serve(lis)
	if err != nil {
		panic(err)
	}
}
