package service

import (
	"time"

	authpb "github.com/CAATHARSIS/music-service/api/gen/auth"
	"github.com/CAATHARSIS/music-service/internal/auth/models"
)

func convertUserToProto(user *models.User) *authpb.User {
	return &authpb.User{
		Id: user.ID,
		Username: user.UserName,
		Email: user.Email,
		Role: user.Role,
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
		UpdatedAt: user.UpdatedAt.Format(time.RFC3339),
	}
}

func convertUserToProfileProto(user *models.User) *authpb.UserProfile {
	pb := &authpb.UserProfile{
		Id: user.ID,
		Username: user.UserName,
		Email: user.Email,
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
	}
	if user.AvatarURL != nil {
		pb.AvatarUrl = user.AvatarURL
	}
	return pb
}