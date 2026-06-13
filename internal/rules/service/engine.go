package service

import (
	"context"
	"fmt"
	"strings"

	catalogpb "github.com/CAATHARSIS/music-service/api/gen/catalog"
	commonpb "github.com/CAATHARSIS/music-service/api/gen/common"
	"github.com/CAATHARSIS/music-service/internal/rules/models"
)

func (s *RuleEngineService) executeRule(ctx context.Context, rule *models.Rule) ([]string, error) {
	if rule == nil {
		return nil, fmt.Errorf("rule cannot be nil")
	}

	var genreIDs []string
	if len(rule.Condition.Genres) > 0 {
		genreResp, err := s.catalogClient.ListGenres(ctx, &catalogpb.ListGenresRequest{})
		if err != nil {
			return nil, fmt.Errorf("get genres: %w", err)
		}
		genreMap := make(map[string]string)
		for _, g := range genreResp.Genres {
			genreMap[strings.ToLower(g.Name)] = g.Id
		}
		for _, genreName := range rule.Condition.Genres {
			if id, ok := genreMap[strings.ToLower(genreName)]; ok {
				genreIDs = append(genreIDs, id)
			}
		}
	}

	req := &catalogpb.ListTracksRequest{
		Pagination: &commonpb.PaginationRequest{
			Page:     1,
			PageSize: int32(rule.TrackLimit),
		},
		GenreIds:  genreIDs,
		SortBy:    catalogpb.TrackSortBy_TRACK_SORT_BY_UNSPECIFIED,
		SortOrder: catalogpb.SortOrder_SORT_ORDER_DESC,
	}

	if rule.Condition.YearFrom != nil {
		req.YearFrom = rule.Condition.YearFrom
	}
	if rule.Condition.YearTo != nil {
		req.YearTo = rule.Condition.YearTo
	}

	if rule.Condition.MinPlaysCount != nil {
		req.YearFrom = rule.Condition.MinPlaysCount
	}
	if rule.Condition.MaxPlaysCount != nil {
		req.YearTo = rule.Condition.MaxPlaysCount
	}

	if rule.Condition.MinDuration != nil {
		req.YearFrom = rule.Condition.MinDuration
	}
	if rule.Condition.MaxDuration != nil {
		req.YearTo = rule.Condition.MaxDuration
	}

	resp, err := s.catalogClient.ListTracks(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("catalog list tracks: %w", err)
	}

	trackIDs := make([]string, len(resp.Tracks))
	for i, t := range resp.Tracks {
		trackIDs[i] = t.Id
	}

	return trackIDs, nil
}
