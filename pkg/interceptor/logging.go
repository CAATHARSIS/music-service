package interceptor

import (
    "context"
    "log/slog"
    "time"

    "google.golang.org/grpc"
)

// Logging возвращает gRPC interceptor для логирования запросов
func Logging(logger *slog.Logger) grpc.UnaryServerInterceptor {
    return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
        start := time.Now()
        resp, err := handler(ctx, req)
        logger.Info("request completed",
            "method", info.FullMethod,
            "duration", time.Since(start),
            "error", err,
        )
        return resp, err
    }
}