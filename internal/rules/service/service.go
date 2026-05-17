package service

import (
	"context"
	"errors"
	"log/slog"

	catalogpb "github.com/CAATHARSIS/music-service/api/gen/catalog"
	commonpb "github.com/CAATHARSIS/music-service/api/gen/common"
	playlistpb "github.com/CAATHARSIS/music-service/api/gen/playlist"
	rulespb "github.com/CAATHARSIS/music-service/api/gen/rules"
	"github.com/CAATHARSIS/music-service/internal/rules/models"
	"github.com/CAATHARSIS/music-service/internal/rules/repository"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type RuleEngineService interface {
	CreateRule(ctx context.Context, req *rulespb.CreateRuleRequest) (*rulespb.Rule, error)
	GetRule(ctx context.Context, req *rulespb.GetRuleRequest) (*rulespb.Rule, error)
	ListUserRules(ctx context.Context, req *rulespb.ListUserRulesRequest) (*rulespb.ListRulesResponse, error)
}

type ruleEngineService struct {
	rulespb.UnimplementedRuleServiceServer
	repo           repository.Repository
	catalogClient  catalogpb.CatalogServiceClient
	playlistClient playlistpb.PlaylistServiceClient
	log            *slog.Logger
}

func NewRuleEngineService(
	repo repository.Repository,
	clatalogClient catalogpb.CatalogServiceClient,
	playlistClient playlistpb.PlaylistServiceClient,
	log *slog.Logger,
) RuleEngineService {
	return &ruleEngineService{
		repo: repo,
		catalogClient: clatalogClient,
		playlistClient: playlistClient,
		log: log,
	}
}

func (s *ruleEngineService) CreateRule(ctx context.Context, req *rulespb.CreateRuleRequest) (*rulespb.Rule, error) {
	if req.UserId == "" || req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id and name required")
	}

	rule := &models.Rule{
		UserID:     req.UserId,
		Name:       req.Name,
		Condition:  convertConditionFromProto(req.Condition),
		TrackLimit: int(req.Limit),
		IsActive:   true,
	}
	if req.Limit == 0 {
		rule.TrackLimit = 50
	}
	if req.CronSchedule != "" {
		rule.CronSchedule = &req.CronSchedule
	}

	if err := s.repo.CreateRule(ctx, rule); err != nil {
		return nil, status.Error(codes.Internal, "create failed")
	}

	return convertRuleToProto(rule), nil
}

func (s *ruleEngineService) GetRule(ctx context.Context, req *rulespb.GetRuleRequest) (*rulespb.Rule, error) {
	rule, err := s.repo.GetRule(ctx, req.RuleId)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "not found")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}
	return convertRuleToProto(rule), nil
}

func (s *ruleEngineService) ListUserRules(ctx context.Context, req *rulespb.ListUserRulesRequest) (*rulespb.ListRulesResponse, error) {
	rules, err := s.repo.ListUserRules(ctx, req.UserId)
	if err != nil {
		return nil, status.Error(codes.Internal, "list failed")
	}

	pbRules := make([]*rulespb.Rule, len(rules))
	for i, r := range rules {
		pbRules[i] = convertRuleToProto(r)
	}
	return &rulespb.ListRulesResponse{Rules: pbRules}, nil
}

func (s *ruleEngineService) ExecuteRule(ctx context.Context, req *rulespb.ExecuteRuleRequest) (*rulespb.ExecuteRuleResponse, error) {
	rule, err := s.repo.GetRule(ctx, req.RuleId)
	if err != nil {
		return nil, status.Error(codes.NotFound, "rule not found")
	}

	trackIDs, err := s.executeRule(ctx, rule)
	if err != nil {
		return nil, status.Error(codes.Internal, "execution failed")
	}

	playlistResp, err := s.playlistClient.CreatePlaylist(ctx, &playlistpb.CreatePlaylistRequest{
		UserId: rule.UserID,
		Name:   rule.Name,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, "create playlist failed")
	}

	for _, trackID := range trackIDs {
		if trackID != "" {
			s.playlistClient.AddTrack(ctx, &playlistpb.AddTrackRequest{
				PlaylistId: playlistResp.Id,
				UserId:     rule.UserID,
				TrackId:    trackID,
			})
		}
	}

	s.repo.MarkExecuted(ctx, rule.ID)

	return &rulespb.ExecuteRuleResponse{
		PlaylistId: playlistResp.Id,
		TrackCount: int32(len(trackIDs)),
	}, nil
}

func (s *ruleEngineService) Health(ctx context.Context, req *commonpb.Empty) (*commonpb.HealthyCheckResponse, error) {
	return &commonpb.HealthyCheckResponse{Status: "SERVING"}, nil
}
