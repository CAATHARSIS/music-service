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
		Page:      page,
		PageSize:  pageSize,
		SortBy:    convertTrackSortBy(req.SortBy),
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

	p := &commonpb.PaginationResponse{}
	if req.Pagination != nil {
		if req.Pagination.Page > 0 {
			p.Page = req.Pagination.Page
		}
		if req.Pagination.PageSize > 0 {
			p.PageSize = req.Pagination.PageSize
		}
	}

	return &catalogpb.ListTracksResponse{
		Tracks:     tracks,
		Pagination: p,
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
		Title:        req.Title,
		Duration:     int(req.Duration),
		Year:         int(req.Year),
		ArtistID:     req.ArtistId,
		FileID:       req.FileId,
		GenreIDs:     req.GenreIds,
		CoverImageID: req.CoverImageId,
		TrackNumber:  req.TrackNumber,
		Lyrics:       req.Lyrics,
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
		Title:        req.Title,
		Duration:     req.Duration,
		Year:         req.Year,
		ArtistID:     req.ArtistId,
		AlbumID:      req.AlbumId,
		FileID:       req.FileId,
		CoverImageID: req.CoverImageId,
		TrackNumber:  req.TrackNumber,
		Lyrics:       req.Lyrycs,
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
		Limit:         20,
		IncludeArtist: req.IncludeArtist,
		IncludeAlbum:  req.IncludeAlbum,
	}

	p := &commonpb.PaginationResponse{
		PageSize: int32(opts.Limit),
	}
	if req.Pagination != nil && req.Pagination.PageSize > 0 {
		opts.Limit = int(req.Pagination.PageSize)
		p.PageSize = req.Pagination.PageSize
	}

	tracks, err := s.repo.SearchTracks(ctx, req.Query, opts)
	if err != nil {
		s.log.Error("failed to search tracks", "error", err)
		return nil, status.Error(codes.Internal, "internal error")
	}

	pbTracks := convertTracksToProto(tracks)

	return &catalogpb.ListTracksResponse{
		Tracks:     pbTracks,
		Pagination: p,
	}, nil
}

func (s *CatalogService) IncrementTrackPlaysCount(ctx context.Context, req *catalogpb.IncrementPlaysCountRequest) (*commonpb.Empty, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "track id is required")
	}

	incrementBy := int64(1)
	if req.IncrementBy > 0 {
		incrementBy = int64(req.IncrementBy)
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

func (s *CatalogService) GetTracksByIDs(ctx context.Context, ids []string, opts *models.GetTrackOptions) ([]*models.Track, error) {
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

func (s *CatalogService) IncrementPlaysCount(ctx context.Context, req *catalogpb.IncrementPlaysCountRequest) (*commonpb.Empty, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "track id is required")
	}

	incrementBy := int64(req.IncrementBy)
	if incrementBy < 1 {
		incrementBy = 1
	}

	track, err := s.repo.GetTrackByID(ctx, req.Id, nil)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "track not found")
		}
		s.log.Error("get track failed", "id", req.Id, "error", err)
		return nil, status.Error(codes.Internal, "internal error")
	}

	if err := s.repo.IncrementPlays(ctx, req.Id, incrementBy); err != nil {
		s.log.Error("increment plays failed", "id", req.Id, "error", err)
		return nil, status.Error(codes.Internal, "internal error")
	}

	if err := s.repo.UpdateArtistStats(ctx, track.ArtistID, incrementBy); err != nil {
		s.log.Warn("update artist stats failed", "artist_id", track.ArtistID, "error", err)
	}

	return &commonpb.Empty{}, nil
}

// Artist

func (s *CatalogService) GetArtist(ctx context.Context, req *catalogpb.GetArtistRequest) (*catalogpb.Artist, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "artist id is required")
	}

	artist, err := s.repo.GetArtistByID(ctx, req.Id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "artitst not found")
		}
		s.log.Error("get artist failed", "id", req.Id, "error", err)
		return nil, status.Error(codes.Internal, "internal error")
	}

	return convertArtistToProto(artist), nil
}

func (s *CatalogService) ListArtists(ctx context.Context, req *catalogpb.ListArtistsRequest) (*catalogpb.ListArtistsResponse, error) {
	page, pageSize := s.paginationDefaults(
		int(req.GetPagination().GetPage()),
		int(req.GetPagination().GetPageSize()),
	)

	filter := &models.ArtistFilter{
		Page:      page,
		PageSize:  pageSize,
		SortBy:    convertArtistSortBy(req.SortBy),
		SortOrder: convertSortOrder(req.SortOrder),
	}

	if req.Country != nil {
		filter.Country = *req.Country
	}
	if len(req.GenresId) > 0 {
		filter.GenreIDs = req.GenresId
	}

	result, err := s.repo.ListArtists(ctx, filter)
	if err != nil {
		s.log.Error("list artists failed", "error", err)
		return nil, status.Error(codes.Internal, "internal error")
	}

	p := &commonpb.PaginationResponse{}
	if req.Pagination != nil {
		if req.Pagination.Page > 0 {
			p.Page = req.Pagination.Page
		}
		if req.Pagination.PageSize > 0 {
			p.PageSize = req.Pagination.PageSize
		}
	}

	return &catalogpb.ListArtistsResponse{
		Artists:    convertArtistsToProto(result.Artists),
		Pagination: p,
	}, nil
}

func (s *CatalogService) CreateArtist(ctx context.Context, req *catalogpb.CreateArtistRequest) (*catalogpb.Artist, error) {
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}

	params := &models.CreateArtistParams{
		Name:          req.Name,
		Country:       req.Country,
		AvatarImageID: req.AvatarImageId,
		GenreIDs:      req.GenreIds,
	}

	artist, err := s.repo.CreateArtist(ctx, params)
	if err != nil {
		s.log.Error("create artist failed", "error", err)
		return nil, status.Error(codes.Internal, "internal error")
	}

	return convertArtistToProto(artist), nil
}

func (s *CatalogService) UpdateArtist(ctx context.Context, req *catalogpb.UpdateArtistRequest) (*catalogpb.Artist, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "artist id is required")
	}

	params := &models.UpdateArtistParams{
		Name:          req.Name,
		Country:       req.Country,
		AvatarImageID: req.AvatarImageId,
	}

	if len(req.GenreIds) > 0 {
		params.GenreIDs = &req.GenreIds
	}

	artist, err := s.repo.UpdateArtist(ctx, req.Id, params)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "artist not found")
		}
		s.log.Error("update artist failed", "id", req.Id, "error", err)
		return nil, status.Error(codes.Internal, "internal error")
	}

	return convertArtistToProto(artist), nil
}

func (s *CatalogService) DeleteArtist(ctx context.Context, req *catalogpb.DeleteArtistRequest) (*commonpb.Empty, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "artist id is required")
	}

	if err := s.repo.DeleteArtistByID(ctx, req.Id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "artist not found")
		}
		s.log.Error("delete artist failed", "id", req.Id, "error", err)
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &commonpb.Empty{}, nil
}

func (s *CatalogService) SearchArtists(ctx context.Context, req *catalogpb.SearchArtistsRequest) (*catalogpb.ListArtistsResponse, error) {
	if req.Query == "" {
		return &catalogpb.ListArtistsResponse{}, nil
	}

	limit := 20
	p := &commonpb.PaginationResponse{
		PageSize: 20,
	}
	if req.Pagination != nil && req.Pagination.PageSize > 0 {
		limit = int(req.Pagination.PageSize)
		p.PageSize = req.Pagination.PageSize
	}

	artists, err := s.repo.SearchArtists(ctx, req.Query, limit)
	if err != nil {
		s.log.Error("search artists failed", "query", req.Query, "error", err)
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &catalogpb.ListArtistsResponse{
		Artists:    convertArtistsToProto(artists),
		Pagination: p,
	}, nil
}

func (s *CatalogService) GetArtistTracks(ctx context.Context, req *catalogpb.GetArtistTracksRequest) (*catalogpb.ListTracksResponse, error) {
	if req.ArtistId == "" {
		return nil, status.Error(codes.InvalidArgument, "artist_id is required")
	}

	limit := 20
	p := &commonpb.PaginationResponse{
		PageSize: int32(limit),
	}
	if req.Pagination != nil && req.Pagination.PageSize > 0 {
		limit = int(req.Pagination.PageSize)
		p.PageSize = req.Pagination.PageSize
	}

	tracks, err := s.repo.GetArtistTracks(ctx, req.ArtistId, limit)
	if err != nil {
		s.log.Error("get artist tracks failed", "artist_id", req.ArtistId, "error", err)
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &catalogpb.ListTracksResponse{
		Tracks:     convertTracksToProto(tracks),
		Pagination: p,
	}, nil
}

func (s *CatalogService) GetArtistAlbums(ctx context.Context, req *catalogpb.GetArtistAlbumsRequest) (*catalogpb.ListAlbumsResponse, error) {
	if req.ArtistId == "" {
		return nil, status.Error(codes.InvalidArgument, "artist_id is required")
	}

	albums, err := s.repo.GetArtistAlbums(ctx, req.ArtistId)
	if err != nil {
		s.log.Error("get artist albums failed", "artist_id", req.ArtistId, "error", err)
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &catalogpb.ListAlbumsResponse{
		Albums: convertAlbumsToproto(albums),
	}, nil
}

func (s *CatalogService) GetArtistsByIDs(ctx context.Context, req *catalogpb.GetArtistsByIDsRequest) (*catalogpb.ListArtistsResponse, error) {
	if len(req.Ids) == 0 {
		return &catalogpb.ListArtistsResponse{}, nil
	}

	artists, err := s.repo.GetArtistByIDs(ctx, req.Ids)
	if err != nil {
		s.log.Error("get artists by ids failed", "count", len(req.Ids), "error", err)
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &catalogpb.ListArtistsResponse{
		Artists: convertArtistsToProto(artists),
	}, nil
}

// Albums

func (s *CatalogService) GetAlbum(ctx context.Context, req *catalogpb.GetAlbumRequest) (*catalogpb.Album, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "album id is required")
	}

	album, err := s.repo.GetAlbumByID(ctx, req.Id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "album not found")
		}
		s.log.Error("get album failed", "id", req.Id, "error", err)
		return nil, status.Error(codes.Internal, "internal error")
	}

	return convertAlbumToProto(album), nil
}

func (s *CatalogService) ListAlbums(ctx context.Context, req *catalogpb.ListAlbumsRequest) (*catalogpb.ListAlbumsResponse, error) {
	page, pageSize := s.paginationDefaults(
		int(req.GetPagination().GetPage()),
		int(req.GetPagination().GetPageSize()),
	)

	filter := &models.AlbumFilter{
		Page:      page,
		PageSize:  pageSize,
		SortBy:    convertAlbumSortBy(req.SortBy),
		SortOrder: convertSortOrder(req.SortOrder),
	}

	if req.ArtistId != nil {
		filter.ArtistID = *req.ArtistId
	}
	if req.YearFrom != nil {
		filter.YearFrom = int(*req.YearFrom)
	}
	if req.YearTo != nil {
		filter.YearTo = int(*req.YearTo)
	}
	if len(req.GenresId) > 0 {
		filter.GenreIDs = req.GenresId
	}

	if req.Type != nil {
		filter.AlbumType = convertAlbumTypeFromProto(*req.Type)
	} else {
		filter.AlbumType = models.AlbumTypeUnspecified
	}

	result, err := s.repo.ListAlbums(ctx, filter)
	if err != nil {
		s.log.Error("list albums failed", "error", err)
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &catalogpb.ListAlbumsResponse{
		Albums: convertAlbumsToproto(result.Albums),
		Pagination: &commonpb.PaginationResponse{
			Page:     int32(result.Page),
			PageSize: int32(result.PageSize),
		},
	}, nil
}

func (s *CatalogService) CreateAlbum(ctx context.Context, req *catalogpb.CreateAlbumRequest) (*catalogpb.Album, error) {
	if req.Title == "" {
		return nil, status.Error(codes.InvalidArgument, "title is required")
	}
	if req.ArtistId == "" {
		return nil, status.Error(codes.InvalidArgument, "artist_id is required")
	}

	params := &models.CreateAlbumParams{
		Title:        req.Title,
		Year:         int(req.Year),
		ArtistID:     req.ArtistId,
		AlbumType:    convertAlbumTypeFromProto(req.Type),
		CoverImageID: req.CoverImageId,
		GenresIDs:    req.GenreIds,
	}

	album, err := s.repo.CreateAlbum(ctx, params)
	if err != nil {
		s.log.Error("create album failed", "error", err)
		return nil, status.Error(codes.Internal, "internal error")
	}

	return convertAlbumToProto(album), nil
}

func (s *CatalogService) UpdateAlbum(ctx context.Context, req *catalogpb.UpdateAlbumRequest) (*catalogpb.Album, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "album id is required")
	}

	albumType := convertAlbumTypeFromProto(req.GetType())

	params := &models.UpdateAlbumParams{
		Title:        req.Title,
		Year:         req.Year,
		ArtistID:     req.ArtistId,
		CoverImageID: req.CoverImageId,
		AlbumType:    &albumType,
	}

	if len(req.GenreIds) > 0 {
		params.GenreIDs = req.GenreIds
	}

	album, err := s.repo.UpdateAlbum(ctx, req.Id, params)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "album not found")
		}
		s.log.Error("update album failed", "id", req.Id, "error", err)
		return nil, status.Error(codes.Internal, "internal error")
	}

	return convertAlbumToProto(album), nil
}

func (s *CatalogService) DeleteAlbum(ctx context.Context, req *catalogpb.DeleteAlbumRequest) (*commonpb.Empty, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "album id is required")
	}

	if err := s.repo.DeleteAlbum(ctx, req.Id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "album not found")
		}
		s.log.Error("delete album failed", "id", req.Id, "error", err)
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &commonpb.Empty{}, nil
}

func (s *CatalogService) GetAlbumWithTracks(ctx context.Context, req *catalogpb.GetAlbumTracksRequest) (*catalogpb.AlbumWithTracks, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "album_id is required")
	}

	albumWithTracks, err := s.repo.GetAlbumWithTracksByID(ctx, req.Id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "album not found")
		}
		s.log.Error("get album tracks failed", "album_id", req.Id, "error", err)
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &catalogpb.AlbumWithTracks{
		Album:  convertAlbumToProto(albumWithTracks.Album),
		Tracks: convertTracksToProto(albumWithTracks.Tracks),
		Genres: convertGenresToProto(albumWithTracks.Genres),
	}, nil
}

func (s *CatalogService) SearchAlbums(ctx context.Context, req *catalogpb.SearchAlbumsRequest) (*catalogpb.ListAlbumsResponse, error) {
	if req.Query == "" {
		return &catalogpb.ListAlbumsResponse{}, nil
	}

	opts := &models.SearchAlbumsOptions{
		Limit:         20,
		IncludeArtist: req.IncludeArtist,
	}

	if req.Pagination != nil && req.Pagination.PageSize > 0 {
		opts.Limit = int(req.Pagination.PageSize)
	}

	albums, err := s.repo.SearchAlbums(ctx, req.Query, opts)
	if err != nil {
		s.log.Error("search albums failed", "error", err)
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &catalogpb.ListAlbumsResponse{
		Albums: convertAlbumsToproto(albums),
		Pagination: &commonpb.PaginationResponse{
			PageSize: int32(opts.Limit),
		},
	}, nil
}

// Genres

func (s *CatalogService) GetGenre(ctx context.Context, req *catalogpb.GetGenreRequest) (*catalogpb.Genre, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "genre id is required")
	}

	genre, err := s.repo.GetGenreByID(ctx, req.Id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "genre not found")
		}
		s.log.Error("get genre failed", "id", req.Id, "error", err)
		return nil, status.Error(codes.Internal, "internal error")
	}

	return convertGenreToProto(genre), nil
}

func (s *CatalogService) ListGenres(ctx context.Context, req *catalogpb.ListGenresRequest) (*catalogpb.ListGenresResponse, error) {
	genres, err := s.repo.ListGenres(ctx)
	if err != nil {
		s.log.Error("list genres failed", "error", err)
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &catalogpb.ListGenresResponse{
		Genres: convertGenresToProto(genres),
	}, nil
}

func (s *CatalogService) CreateGenre(ctx context.Context, req *catalogpb.CreateGenreRequest) (*catalogpb.Genre, error) {
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}

	params := &models.CreateGenreParams{
		Name: req.Name,
		Description: req.Description,
	}

	genre, err := s.repo.CreateGenre(ctx, params)
	if err != nil {
		s.log.Error("create genre failed", "error", err)
		return nil, status.Error(codes.Internal, "internal error")
	}

	return convertGenreToProto(genre), nil
}

func (s *CatalogService) GetTrackByGenre(ctx context.Context, req *catalogpb.GetTracksByGenreRequest) (*catalogpb.ListTracksResponse, error) {
	if req.GenreId == "" {
		return nil, status.Error(codes.InvalidArgument, "genre_id is required")
	}

	limit := 20
	if req.Pagination != nil && req.Pagination.PageSize > 0 {
		limit = int(req.Pagination.PageSize)
	}

	tracks, err := s.repo.GetTracksByGenre(ctx, req.GenreId, limit)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "genre not found or no tracks")
		}
		s.log.Error("get tracks by genre failed", "genre_id", req.GenreId, "error", err)
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &catalogpb.ListTracksResponse{
		Tracks: convertTracksToProto(tracks),
		Pagination: &commonpb.PaginationResponse{
			PageSize: int32(limit),
		},
	}, nil
}

// Heath

func (s *CatalogService) Health(ctx context.Context, req *commonpb.Empty) (*commonpb.HealthyCheckResponse, error) {
	return &commonpb.HealthyCheckResponse{
		Status: "Serving",
	}, nil
}

// Helpers

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
