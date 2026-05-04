package service

import (
	"time"

	catalogpb "github.com/CAATHARSIS/music-service/api/gen/catalog"
	"github.com/CAATHARSIS/music-service/internal/catalog/models"
)

func convertTrackToProto(track *models.Track) *catalogpb.Track {
	if track == nil {
		return nil
	}

	pb := &catalogpb.Track {
		Id: track.ID,
		Title: track.Title,
		Duration: int32(track.Duration),
		Year: int32(track.Year),
		FileId: track.FileID,
		PlaysCount: int64(track.PlaysCount),
		CreatedAt: track.CreatedAt.Format(time.RFC3339),
		UpdatedAt: track.UpdatedAt.Format(time.RFC3339),
	}

	if track.CoverImageID != nil {
		pb.CoverImageId = track.CoverImageID
	}
	if track.TrackNumber != nil {
		v := int32(*track.TrackNumber)
		pb.TrackNumber = &v
	}
	if track.Lyrics != nil {
		pb.Lyrics = track.Lyrics
	}

	pb.Artist = convertArtistToProto(track.Artist)
	pb.Album = convertAlbumToProto(track.Album)
	pb.Genres = convertGenresToProto(track.Genres)

	return pb
}

func convertTracksToProto(tracks []*models.Track) []*catalogpb.Track {
	if tracks == nil {
		return nil
	}

	result := make([]*catalogpb.Track, len(tracks))
	for i, track := range tracks {
		result[i] = convertTrackToProto(track)
	}

	return result
}

func convertArtistToProto(artist *models.Artist) *catalogpb.Artist {
	if artist == nil {
		return nil
	}

	pb := &catalogpb.Artist{
		Id: artist.ID,
		Name: artist.Name,
		TotalPlays: artist.TotalPlays,
		CreatedAt: artist.CreatedAt.Format(time.RFC3339),
		UpdatedAt: artist.UpdatedAt.Format(time.RFC3339),
	}

	if artist.Country != nil {
		pb.Country = artist.Country
	}
	if artist.AvatarImageID != nil {
		pb.AvatarImageId = artist.AvatarImageID
	}

	pb.Genres = convertGenresToProto(artist.Genres)

	return pb
}

func convertArtistsToProto(artists []*models.Artist) []*catalogpb.Artist {
	if artists == nil {
		return nil
	}

	result := make([]*catalogpb.Artist, len(artists))
	for i, artist := range artists {
		result[i] = convertArtistToProto(artist)
	}

	return result
}

func convertAlbumToProto(album *models.Album) *catalogpb.Album {
	if album == nil {
		return nil
	}

	pb := &catalogpb.Album{
		Id: album.ID,
		Title: album.Title,
		Year: album.Year,
		Type: convertAlbumTypeToProto(album.AlbumType),
		CreatedAt: album.CreatedAt.Format(time.RFC3339),
		UpdatedAt: album.UpdatedAt.Format(time.RFC3339),
	}

	if album.CoverImageID != nil {
		pb.CoverImageId = album.CoverImageID
	}

	pb.Artist = convertArtistToProto(album.Artist)
	pb.Genres = convertGenresToProto(album.Genres)

	return pb
}

func convertAlbumsToproto(albums []*models.Album) []*catalogpb.Album {
	if albums == nil {
		return nil
	}

	result := make([]*catalogpb.Album, len(albums))
	for i, album := range albums {
		result[i] = convertAlbumToProto(album)
	}

	return result
}

func convertGenreToProto(genre *models.Genre) *catalogpb.Genre {
	if genre == nil {
		return nil
	}

	pb := &catalogpb.Genre{
		Id: genre.ID,
		Name: genre.Name,
	}

	if genre.Description != nil {
		pb.Description = genre.Description
	}

	return pb
}

func convertGenresToProto(genres []*models.Genre) []*catalogpb.Genre {
	if genres == nil {
		return nil
	}

	result := make([]*catalogpb.Genre, len(genres))
	for i, genre := range genres {
		result[i] = convertGenreToProto(genre)
	}

	return result
}

func convertAlbumTypeToProto(albumType models.AlbumType) catalogpb.AlbumType {
	switch albumType {
	case models.AlbumTypeAlbum:
		return catalogpb.AlbumType_ALBUM_TYPE_ALBUM
	case models.AlbumTypeEP:
		return catalogpb.AlbumType_ALBUM_TYPE_EP
	case models.AlbumTypeSingle:
		return catalogpb.AlbumType_ALBUM_TYPE_SINLGE
	default:
		return catalogpb.AlbumType_ALBUM_TYPE_UNSPECIFIED
	}
}

func convertAlbumTypeFromProto(pbType catalogpb.AlbumType) models.AlbumType {
	switch pbType {
	case catalogpb.AlbumType_ALBUM_TYPE_ALBUM:
		return models.AlbumTypeAlbum
	case catalogpb.AlbumType_ALBUM_TYPE_EP:
		return models.AlbumTypeEP
	case catalogpb.AlbumType_ALBUM_TYPE_SINLGE:
		return models.AlbumTypeSingle
	default:
		return models.AlbumTypeUnspecified
	}
}

func convertTrackSortBy(pbSort catalogpb.TrackSortBy) models.TrackSortBy {
	switch pbSort {
	case catalogpb.TrackSortBy_TRACK_SORT_BY_TITLE:
		return models.SortByTitle
	case catalogpb.TrackSortBy_TRACK_SORT_BY_ARTIST:
		return models.SortByArtist
	case catalogpb.TrackSortBy_TRACK_SORT_BY_YEAR:
		return models.SortByYear
	default:
		return models.SortByTitle
	}
}

func convertArtistSortBy(pbSort catalogpb.ArtistSortBy) models.ArtistSortBy {
	switch pbSort {
	case catalogpb.ArtistSortBy_ARTIST_SORT_BY_NAME:
		return models.SortArtistsByName
	case catalogpb.ArtistSortBy_ARTIST_SORT_BY_TOTAL_PLAYS:
		return models.SortArtistsByPlays
	default:
		return models.SortArtistsByName
	}
}

func convertAlbumSortBy(pbSort catalogpb.AlbumSortBy) models.AlbumSortBy {
	switch pbSort {
	case catalogpb.AlbumSortBy_ALBUM_SORT_BY_TITLE:
		return models.SortAlbumByTitle
	case catalogpb.AlbumSortBy_ALBUM_SORT_BY_ARTIST:
		return models.SortAlbumByArtist
	case catalogpb.AlbumSortBy_ALBUM_SORT_BY_YEAR:
		return models.SortAlbumByYear
	default:
		return models.SortAlbumByTitle
	}
}

func convertSortOrder(pbOrder catalogpb.SortOrder) models.SortOrder {
	switch pbOrder {
	case catalogpb.SortOrder_SORT_ORDER_ASC:
		return models.SortOrderAsc
	case catalogpb.SortOrder_SORT_ORDER_DESC:
		return models.SortOrderDesc
	default:
		return ""
	}
}