package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"time"

	authpb "github.com/CAATHARSIS/music-service/api/gen/auth"
	commonpb "github.com/CAATHARSIS/music-service/api/gen/common"
	filepb "github.com/CAATHARSIS/music-service/api/gen/file"
	"github.com/CAATHARSIS/music-service/internal/auth/config"
	"github.com/CAATHARSIS/music-service/internal/auth/models"
	"github.com/CAATHARSIS/music-service/internal/auth/repository"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthService struct {
	authpb.UnimplementedAuthServiceServer
	repo       repository.Repository
	fileClient filepb.FileServiceClient
	cfg        *config.Config
	log        *slog.Logger
}

func NewAuthService(repo repository.Repository, fileClient filepb.FileServiceClient, cfg *config.Config, log *slog.Logger) *AuthService {
	return &AuthService{
		repo: repo,
		fileClient: fileClient,
		cfg:  cfg,
		log:  log,
	}
}

func (s *AuthService) Register(ctx context.Context, req *authpb.RegisterRequest) (*authpb.AuthResponse, error) {
	if req.Username == "" || req.Email == "" || req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "username, email and password are required")
	}

	user, err := s.repo.CreateUser(ctx, req.Username, req.Email, req.Password)
	if err != nil {
		if errors.Is(err, repository.ErrAlreadyExists) {
			return nil, status.Error(codes.AlreadyExists, "user already exists")
		}
		s.log.Error("create user failed", "error", err)
		return nil, status.Error(codes.Internal, "internal error")
	}

	token, err := s.generateAccessToken(user)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to generate token: %v", err))
	}

	refreshToken, err := s.generateRefreshToken(ctx, user)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to generate refresh token")
	}

	return &authpb.AuthResponse{
		AccessToken:  token,
		RefreshToken: refreshToken,
		ExpiresIn:    s.getExpirationSeconds(),
		User:         convertUserToProto(user),
	}, nil
}

func (s *AuthService) Login(ctx context.Context, req *authpb.LoginRequest) (*authpb.AuthResponse, error) {
	if req.Email == "" || req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "email and password is required")
	}

	user, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}

	if !s.repo.VerifyPassword(user, req.Password) {
		return nil, status.Error(codes.Unauthenticated, "invalid password")
	}

	token, err := s.generateAccessToken(user)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to generate token")
	}

	refreshToken, err := s.generateRefreshToken(ctx, user)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to generate referesh token")
	}

	return &authpb.AuthResponse{
		AccessToken:  token,
		RefreshToken: refreshToken,
		ExpiresIn:    s.getExpirationSeconds(),
		User:         convertUserToProto(user),
	}, nil
}

func (s AuthService) ValidateToken(ctx context.Context, req *authpb.ValidateTokenRequest) (*authpb.ValidateTokenResponse, error) {
	token, err := jwt.Parse(req.Token, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.cfg.JWTSecret), nil
	})

	if err != nil || !token.Valid {
		return &authpb.ValidateTokenResponse{Valid: false}, nil
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return &authpb.ValidateTokenResponse{Valid: false}, nil
	}

	userID, _ := claims["sub"].(string)
	role, _ := claims["role"].(string)

	return &authpb.ValidateTokenResponse{
		Valid:  true,
		UserId: userID,
		Role:   role,
	}, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, req *authpb.RefreshTokenRequest) (*authpb.AuthResponse, error) {
	hash := hashToken(req.RefreshToken)

	rt, err := s.repo.GetRefreshToken(ctx, hash)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, status.Error(codes.Unauthenticated, "invalid refresh token")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}

	s.repo.DeleteRefreshToken(ctx, rt.ID)

	user, err := s.repo.GetUserByID(ctx, rt.UserID)
	if err != nil {
		return nil, status.Error(codes.Internal, "user not found")
	}

	token, _ := s.generateAccessToken(user)
	newRefreshToken, _ := s.generateRefreshToken(ctx, user)

	return &authpb.AuthResponse{
		AccessToken:  token,
		RefreshToken: newRefreshToken,
		ExpiresIn:    s.getExpirationSeconds(),
		User:         convertUserToProto(user),
	}, nil
}

func (s *AuthService) GetProfile(ctx context.Context, req *authpb.GetProfileRequest) (*authpb.UserProfile, error) {
	user, err := s.repo.GetUserByID(ctx, req.UserId)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}

	return convertUserToProfileProto(user), nil
}

func (s *AuthService) GetAvatarURL(ctx context.Context, req *authpb.GetAvatarURLRequest) (*authpb.AvatarURLResponse, error) {
	resp, err := s.fileClient.GetDownloadURL(ctx, &filepb.GetDownloadURLRequest{
		FileId: req.AvatarImageId,
		ExpirySeconds: 86400,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get avatar URL")
	}

	return &authpb.AvatarURLResponse{
		Url: resp.Url,
		ExpiresAt: resp.ExpiresAt,
	}, nil
}

func (s *AuthService) Health(ctx context.Context, req *commonpb.Empty) (*commonpb.HealthyCheckResponse, error) {
	return &commonpb.HealthyCheckResponse{Status: "SERVING"}, nil
}

func (s *AuthService) generateAccessToken(user *models.User) (string, error) {
	duration, _ := time.ParseDuration(s.cfg.JWTExpiration)

	claims := jwt.MapClaims{
		"sub":  user.ID,
		"role": user.Role,
		"exp":  time.Now().Add(duration).Unix(),
		"iat":  time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JWTSecret))
}

func (s *AuthService) generateRefreshToken(ctx context.Context, user *models.User) (string, error) {
	raw := uuid.New().String() + user.ID + time.Now().String()
	hash := hashToken(raw)

	err := s.repo.SaveRefreshToken(ctx, user.ID, hash, time.Now().Add(30*24*time.Hour))
	if err != nil {
		return "", err
	}

	return raw, nil
}

func (s *AuthService) getExpirationSeconds() int64 {
	duration, _ := time.ParseDuration(s.cfg.JWTExpiration)
	return int64(duration.Seconds())
}

func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}
