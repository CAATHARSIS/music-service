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
		req.YearFrom = toPtr(int32(*rule.Condition.YearFrom))
	}
	if rule.Condition.YearTo != nil {
		req.YearTo = toPtr(int32(*rule.Condition.YearTo))
	}

	resp, err := s.catalogClient.ListTracks(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("catalog list tracks: %w", err)
	}

	var trackIDs []string
	for _, t := range resp.Tracks {
		if rule.Condition.MinPlaysCount != nil && t.PlaysCount < int64(*rule.Condition.MinPlaysCount) {
			continue
		}
		if rule.Condition.MinDuration != nil && t.Duration < int32(*rule.Condition.MinDuration) {
			continue
		}
		if rule.Condition.MaxDuration != nil && t.Duration > int32(*rule.Condition.MaxDuration) {
			continue
		}
		trackIDs = append(trackIDs, t.Id)
	}

	return trackIDs, nil
}
