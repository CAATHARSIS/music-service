package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	catalogpb "github.com/CAATHARSIS/music-service/api/gen/catalog"
	commonpb "github.com/CAATHARSIS/music-service/api/gen/common"
	"github.com/CAATHARSIS/music-service/internal/catalog/models"
	"github.com/CAATHARSIS/music-service/internal/catalog/repository"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CatalogService struct {
	catalogpb.UnimplementedCatalogServiceServer
	repo repository.Repository
	log  *slog.Logger
}

func NewCatalogService(repo repository.Repository, log *slog.Logger) *CatalogService {
	return &CatalogService{
		repo: repo,
		log:  log,
	}
}

// Tracks

func (s *CatalogService) GetTrack(ctx context.Context, req *catalogpb.GetTrackRequest) (*catalogpb.Track, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "track id is required")
	}

	opts := &models.GetTrackOptions{
		IncludeArtist: req.IncludeArtist,
		IncludeAlbum:  req.IncludeAlbum,
		IncludeGenres: req.IncludeGenres,
	}

	track, err := s.repo.GetTrackByID(ctx, req.Id, opts)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "track not found")
		}
		s.log.Error("failed to get track by id", "error", err)
		return nil, status.Error(codes.Internal, "internal error")
	}

	return convertTrackToProto(track), nil
}

func (s *CatalogService) ListTracks(ctx context.Context, req *catalogpb.ListTracksRequest) (*catalogpb.ListTracksResponse, error) {
	page, pageSize := s.paginationDefaults(
		int(req.GetPagination().GetPage()),
		int(req.GetPagination().GetPageSize()),
	)

	filter := &models.TrackFilter{
		Page: page,
		PageSize: pageSize,
		SortBy: convertTrackSortBy(req.SortBy),
		SortOrder: convertSortOrder(req.SortOder),
	}

	if req.ArtistId != nil {
		filter.ArtistID = *req.ArtistId
	}
	if req.AlbumId != nil {
		filter.AlbumID = *req.AlbumId
	}
	if req.YearFrom != nil {
		filter.YearFrom = int(*req.YearFrom)
	}
	if req.YearTo != nil {
		filter.YearTo = int(*req.YearTo)
	}
	if len(req.GenreIds) > 0 {
		filter.GenreIDs = req.GenreIds
	}

	result, err := s.repo.ListTracks(ctx, filter)
	if err != nil {
		s.log.Error("failed to list tracks", "error", err)
		return nil, status.Error(codes.Internal, "internal error")
	}

	tracks := convertTracksToProto(result.Tracks)

	return &catalogpb.ListTracksResponse{
		Track: tracks,
		Pagination: &commonpb.PaginationResponse{
			Page: int32(result.Page),
			PageSize: int32(result.PageSize),
		},
	}, nil
}

func (s *CatalogService) CreateTrack(ctx context.Context, req *catalogpb.CreateTrackRequest) (*catalogpb.Track, error) {
	if req.Title == "" {
		return nil, status.Error(codes.InvalidArgument, "title is required")
	}
	if req.ArtistId == "" {
		return nil, status.Error(codes.InvalidArgument, "artist_id is required")
	}
	if req.FileId == "" {
		return nil, status.Error(codes.InvalidArgument, "file_id is required")
	}

	params := &models.CreateTrackParams{
		Title: req.Title,
		Duration: int(req.Duration),
		Year: int(req.Year),
		ArtistID: req.ArtistId,
		FileID: req.FileId,
		GenreIDs: req.GenreIds,
		CoverImageID: req.CoverImageId,
		TrackNumber: req.TrackNumber,
		Lyrics: req.Lyrics,
	}

	if req.AlbumId != "" {
		params.AlbumID = &req.AlbumId
	}

	track, err := s.repo.CreateTrack(ctx, params)
	if err != nil {
		s.log.Error("failed to create track", "error", err)
		return nil, status.Error(codes.Internal, "internal error")
	}

	return convertTrackToProto(track), nil
}

func (s *CatalogService) UpdateTrack(ctx context.Context, req *catalogpb.UpdateTrackRequest) (*catalogpb.Track, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "track id is required")
	}

	params := &models.UpdateTrackParams{
		Title: req.Title,
		Duration: req.Duration,
		Year: req.Year,
		ArtistID: req.ArtistId,
		AlbumID: req.AlbumId,
		FileID: req.FileId,
		CoverImageID: req.CoverImageId,
		TrackNumber: req.TrackNumber,
		Lyrics: req.Lyrycs,
	}

	if len(req.GenresId) > 0 {
		params.GenreIDs = &req.GenresId
	}

	track, err := s.repo.UpdateTrack(ctx, req.Id, params)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "track not found")
		}
		s.log.Error("fatiled to update", "error", err)
		return nil, status.Error(codes.Internal, "internal error")
	}

	return convertTrackToProto(track), nil
}

func (s *CatalogService) DeleteTrack(ctx context.Context, req *catalogpb.DeleteTrackRequest) (*commonpb.Empty, error) {
	if req.TrackId == "" {
		return nil, status.Error(codes.InvalidArgument, "track_id is required")
	}

	if err := s.repo.DeleteTrackByID(ctx, req.TrackId); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "track not found")
		}
		s.log.Error("failed to delete track", "error", err)
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &commonpb.Empty{}, nil
}

func (s *CatalogService) SearchTracks(ctx context.Context, req *catalogpb.SearchTrackRequest) (*catalogpb.ListTracksResponse, error) {
	if req.Query == "" {
		return &catalogpb.ListTracksResponse{}, nil
	}

	opts := &models.SearchTracksOptions{
		Limit: 20,
		IncludeArtist: req.IncludeArtist,
		IncludeAlbum: req.IncludeAlbum,
	}

	if req.Pagination != nil && req.Pagination.PageSize > 0 {
		opts.Limit = int(req.Pagination.PageSize)
	}

	tracks, err := s.repo.SearchTracks(ctx, req.Query, opts)
	if err != nil {
		s.log.Error("failed to search tracks", "error", err)
		return nil, status.Error(codes.Internal, "internal error")
	}

	pbTracks := convertTracksToProto(tracks)

	return &catalogpb.ListTracksResponse{
		Track: pbTracks,
		Pagination: &commonpb.PaginationResponse{
			Page: req.Pagination.Page,
			PageSize: req.Pagination.PageSize,
		},
	}, nil
}

func (s *CatalogService) IncrementPlays(ctx context.Context, req *catalogpb.IncrementPlaysCountRequest) (*commonpb.Empty, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "track id is required")
	}

	incrementBy := int64(1)
	if req.IncrementBy > 0 {
		incrementBy= int64(req.IncrementBy)
	}

	if err := s.repo.IncrementPlays(ctx, req.Id, incrementBy); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "track not found")
		}
		s.log.Error("failed to increment plays", "error", err)
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &commonpb.Empty{}, nil
}

func (s *CatalogService) GetTrackByIDs(ctx context.Context, ids []string, opts *models.GetTrackOptions) ([]*models.Track, error) {
	if len(ids) == 0 {
		return []*models.Track{}, nil
	}

	tracks, err := s.repo.GetTracksByIDs(ctx, ids, opts)
	if err != nil {
		s.log.Error("get tracks by ids failed", "ids_count", len(ids), "error", err)
		return nil, fmt.Errorf("get tracks by ids: %w", err)
	}

	return tracks, err
}

func (s *CatalogService) paginationDefaults(page, pageSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	return page, pageSize
}