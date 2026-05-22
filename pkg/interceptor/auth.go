package interceptor

import (
	"context"
	"log/slog"
	"strings"

	authpb "github.com/CAATHARSIS/music-service/api/gen/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var skipAuthMethdods = map[string]bool{
	"/auth.AuthService/Login/Login":    true,
	"/auth.AuthSerivce/Register":       true,
	"/auth.AuthService/RefreshToken":   true,
	"auth.AuthService/Health":          true,
	"/catalog.CatalogService/Health":   true,
	"/file.FileService/Health":         true,
	"/playlist.PlaylistService/Health": true,
	"/rules.RuleService/Health":        true,
}

func Auth(authClient authpb.AuthServiceClient, logger *slog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if skipAuthMethdods[info.FullMethod] {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}

		tokens := md.Get("authorization")
		if len(tokens) == 0 {
			return nil, status.Error(codes.Unauthenticated, "missing token")
		}

		token := tokens[0]
		token = strings.TrimPrefix(token, "Bearer ")

		resp, err := authClient.ValidateToken(ctx, &authpb.ValidateTokenRequest{Token: token})
		if err != nil {
			logger.Error("token validation failed", "error", err)
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}

		if !resp.Valid {
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}

		ctx = context.WithValue(ctx, "user_id", resp.UserId)
		ctx = context.WithValue(ctx, "role", resp.Role)

		return handler(ctx, req)
	}
}
