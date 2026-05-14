package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	catalogpb "github.com/CAATHARSIS/music-service/api/gen/catalog"
	commonpb "github.com/CAATHARSIS/music-service/api/gen/common"
	playlistpb "github.com/CAATHARSIS/music-service/api/gen/playlist"
	"github.com/CAATHARSIS/music-service/internal/playlist/models"
	"github.com/CAATHARSIS/music-service/internal/playlist/repository"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type PlaylistService struct {
	playlistpb.UnimplementedPlaylistServiceServer
	repo          repository.Repository
	catalogClient catalogpb.CatalogServiceClient
	log           *slog.Logger
}

func NewPlaylistService(repo repository.Repository, catalogClient catalogpb.CatalogServiceClient, log *slog.Logger) *PlaylistService {
	return &PlaylistService{
		repo:          repo,
		catalogClient: catalogClient,
		log:           log,
	}
}

func (s *PlaylistService) CreatePlaylist(ctx context.Context, req *playlistpb.CreatePlaylistRequest) (*playlistpb.Playlist, error) {
	if req.UserId == "" || req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id and name are required")
	}

	var pType models.PlaylistType
	switch req.PType {
	case playlistpb.PlaylistType_PLAYLIST_TYPE_FAVORITES:
		pType = models.PlaylistTypeFavoriets
	case playlistpb.PlaylistType_PLAYLIST_TYPE_GENERATED:
		pType = models.PlaylistTypeGenerated
	case playlistpb.PlaylistType_PLAYLIST_TYPE_MANUAL:
		pType = models.PlaylistTypeManual
	default:
		pType = models.PlaylistTypeUnspecified
	}

	p, err := s.repo.CreatePlaylist(ctx, req.UserId, req.Name, req.Description, req.IsPublic, pType)
	if err != nil {
		s.log.Error("create playlist failed", "error", err)
		return nil, status.Error(codes.Internal, "create failed")
	}

	return convertPlaylistToProto(p), nil
}

func (s *PlaylistService) GetPlaylist(ctx context.Context, req *playlistpb.GetPlaylistRequest) (*playlistpb.Playlist, error) {
	if req.PlaylistId == "" {
		return nil, status.Error(codes.InvalidArgument, "playlist_id is required")
	}

	p, err := s.repo.GetPlaylist(ctx, req.PlaylistId)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "not found")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}

	pb := convertPlaylistToProto(p)

	tracks, err := s.repo.GetPlaylistTracks(ctx, req.PlaylistId)
	if err != nil {
		s.log.Warn("failed to load tracks", "error", err)
		return pb, nil
	}

	if len(tracks) == 0 {
		return pb, nil
	}

	tracksIDs := make([]string, len(tracks))
	for i, t := range tracks {
		tracksIDs[i] = t.TrackID
	}

	catalogResp, err := s.catalogClient.ListTracks(ctx, &catalogpb.ListTracksRequest{
		Pagination: &commonpb.PaginationRequest{Page: 1, PageSize: int32(len(tracksIDs))},
	})
	if err != nil {
		s.log.Warn("failed to load track info from catalog", "error", err)
		for _, t := range tracks {
			pb.Tracks = append(pb.Tracks, convertPlaylistTrackToProto(t))
		}
		return pb, nil
	}

	trackMap := make(map[string]*catalogpb.Track)
	for _, track := range catalogResp.Tracks {
		trackMap[track.Id] = track
	}

	for _, t := range tracks {
		pt := convertPlaylistTrackToProto(t)
		if info, ok := trackMap[t.TrackID]; ok {
			pt.TrackInfo = info
		}
		pb.Tracks = append(pb.Tracks, pt)
	}

	return pb, nil
}

func (s *PlaylistService) ListUserPlaylists(ctx context.Context, req *playlistpb.ListUserPlaylistsRequest) (*playlistpb.ListPlaylistsResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	page, pageSize := paginationDefaults(req.GetPagination())
	playlists, err := s.repo.ListUserPlaylists(ctx, req.UserId, page, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, "list failed")
	}

	pbPlaylists := make([]*playlistpb.Playlist, len(playlists))
	for i, p := range playlists {
		pbPlaylists[i] = convertPlaylistToProto(p)
	}

	return &playlistpb.ListPlaylistsResponse{
		Playlists: pbPlaylists,
		Pagination: &commonpb.PaginationResponse{
			Page:     int32(page),
			PageSize: int32(pageSize),
		},
	}, nil
}

func (s *PlaylistService) UpdatePlaylist(ctx context.Context, req *playlistpb.UpdatePlaylistRequest) (*playlistpb.Playlist, error) {
	if req.PlaylistId == "" {
		return nil, status.Error(codes.InvalidArgument, "playlist_id is required")
	}

	p, err := s.repo.UpdatePlaylist(ctx, req.PlaylistId, req.Name, req.Description, req.IsPublic)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "not found")
		}
		return nil, status.Error(codes.Internal, fmt.Sprintf("update failed: %v", err))
	}

	return convertPlaylistToProto(p), nil
}

func (s *PlaylistService) DeletePlaylist(ctx context.Context, req *playlistpb.DeletePlaylistRequest) (*commonpb.Empty, error) {
	if req.PlaylistId == "" {
		return nil, status.Error(codes.InvalidArgument, "playlist_id is required")
	}

	if err := s.repo.DeletePlaylist(ctx, req.PlaylistId); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "not found")
		}
		return nil, status.Error(codes.Internal, "delete failed")
	}

	return &commonpb.Empty{}, nil
}

func (s *PlaylistService) AddTrack(ctx context.Context, req *playlistpb.AddTrackRequest) (*playlistpb.Playlist, error) {
	if req.PlaylistId == "" || req.TrackId == "" {
		return nil, status.Error(codes.InvalidArgument, "playlist_id and track_id are required")
	}

	if err := s.repo.AddTrack(ctx, req.PlaylistId, req.TrackId); err != nil {
		return nil, status.Error(codes.Internal, "add track failed")
	}

	return s.GetPlaylist(ctx, &playlistpb.GetPlaylistRequest{PlaylistId: req.PlaylistId})
}

func (s *PlaylistService) RemoveTrack(ctx context.Context, req *playlistpb.RemoveTrackRequest) (*playlistpb.Playlist, error) {
	if req.PlaylistId == "" || req.TrackId == "" {
		return nil, status.Error(codes.InvalidArgument, "playlist_id and track_id are required")
	}

	if err := s.repo.RemoveTrack(ctx, req.PlaylistId, req.TrackId); err != nil {
		return nil, status.Error(codes.Internal, "remove track failed")
	}

	return s.GetPlaylist(ctx, &playlistpb.GetPlaylistRequest{PlaylistId: req.PlaylistId})
}

func (s *PlaylistService) Health(ctx context.Context, req *commonpb.Empty) (*commonpb.HealthyCheckResponse, error) {
	return &commonpb.HealthyCheckResponse{Status: "SERVING"}, nil
}

func paginationDefaults(pb *commonpb.PaginationRequest) (int, int) {
	page, pageSize := 1, 20
	if pb != nil {
		if pb.Page > 0 {
			page = int(pb.Page)
		}
		if pb.PageSize > 0 && pb.PageSize <= 100 {
			pageSize = int(pb.PageSize)
		}
	}
	return page, pageSize
}
