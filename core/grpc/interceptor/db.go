package interceptor

import (
	"context"
	"google.golang.org/grpc"
)

var DBContext = func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	// TODO 扩展连接器
	return handler(ctx, req)
}
