package auth

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func RequireAdmin(ctx context.Context) error {
	role, ok := ctx.Value("role").(string)
	if !ok || role != "admin" {
		return status.Error(codes.PermissionDenied, "admin access required")
	}
	return nil
}
