package leaderboards

import (
	"context"
	"encoding/json"
	"time"

	"github.com/alee792/teamfit/pkg/tft"
	"github.com/pkg/errors"
)

// Result for a single player's match.
type Result struct {
	MatchID   string    `json:"match_id"`
	StartedAt time.Time `json:"started_at"`
	tft.Participant
}

// GetResultsArgs allows users to query match results.
// Must be paired with additional identifiers, either a Leaderboard or list of Summoners.
type GetResultsArgs struct {
	GameLimit int
	Before    time.Time
	After     time.Time
}

// PUUIDResults returns Summoner's match results with PUUIDs as keys.
type PUUIDResults map[string][]Result

// NameResults returns Summoner's match results with names as keys.
type NameResults map[string][]Result

type Summoner struct {
	tft.Summoner
	tft.LeagueEntry
}

// MarshalJSON hides confidential fields.
func (s *Summoner) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		// Summoner
		ProileIconID  int    `json:"-"`
		Name          string `json:"name"`
		PUUID         string `json:"-"`
		SummonerLevel int    `json:"summonerLevel"`
		AccountID     string `json:"-"`
		ID            string `json:"-"`
		RevisionDate  int    `json:"revisionDate"`
		// LeagueEntry
		Inactive     bool   `json:"inactive"`
		FreshBlood   bool   `json:"freshBlood"`
		Veteran      bool   `json:"veteran"`
		HotStreak    bool   `json:"hotStreak"`
		QueueType    string `json:"queueType"`
		SummonerName string `json:"summonerName,omitempty"`
		Wins         int    `json:"wins"`
		Losses       int    `json:"losses"`
		Rank         string `json:"rank"`
		LeagueID     string `json:"leagueId"`
		Tier         string `json:"tier"`
		SummonerID   string `json:"summonerID"`
		LeaguePoints int    `json:"leaguePoints"`
		tft.MiniSeries
	}{
		Name:          s.Name,
		SummonerLevel: s.SummonerLevel,
		RevisionDate:  s.RevisionDate,
		Inactive:      s.Inactive,
		FreshBlood:    s.FreshBlood,
		Veteran:       s.Veteran,
		HotStreak:     s.HotStreak,
		QueueType:     s.QueueType,
		Wins:          s.Wins,
		Losses:        s.Losses,
		Rank:          s.Rank,
		LeagueID:      s.LeagueID,
		Tier:          s.Tier,
		SummonerID:    s.SummonerID,
		LeaguePoints:  s.LeaguePoints,
		MiniSeries:    s.MiniSeries,
	})
}

// GetResults for a set of Summoners.
func (s *Server) GetResults(ctx context.Context, puuids []string, in *GetResultsArgs) (PUUIDResults, error) {
	if in.GameLimit < 1 {
		in.GameLimit = 1
	}

	// Collate match IDs to query.
	var (
		matches  = make(map[string]*tft.Match)
		matchIDs []string
	)

	for _, id := range puuids {
		out, err := s.API.ListMatches(ctx, &tft.ListMatchesRequest{
			PUUID: id,
		})
		if err != nil {
			return nil, errors.Wrapf(err, "failed to list matches for %s", id)
		}

		matchIDs = out.MatchIDs
		for _, id := range matchIDs {
			matches[id] = nil
		}
	}

	// Prepare results and Leaderboard.
	// Participants are really match results for participants.
	var results PUUIDResults = make(map[string][]Result) // Key = Summoner.PUUID
	for _, id := range puuids {
		results[id] = []Result{}
	}

	// Retrieve matches and collate results.
	for _, id := range matchIDs {
		out, err := s.API.GetMatch(ctx, &tft.GetMatchRequest{
			MatchID: id,
		})
		if err != nil {
			return nil, errors.Wrapf(err, "unable to retrieve match %s", id)
		}

		// Enforce time range.
		gameStart := UnixMS(out.Match.Info.GameTimestamp)
		if (!in.Before.IsZero() && gameStart.After(in.Before)) || (!in.After.IsZero() && gameStart.Before(in.After)) {
			continue
		}

		matches[id] = &out.Match

		// Append match results if player is tracked on Leaderboard.
		for _, p := range out.Match.Info.Participants {
			_, ok := results[p.PUUID]
			if !ok {
				continue
			}

			// Do not append results if a player exceeds the match limit.
			// The assumption is that listed match IDs are returned in chronological order.
			if len(results[p.PUUID]) >= in.GameLimit {
				break
			}

			results[p.PUUID] = append(results[p.PUUID], Result{
				MatchID:     id,
				StartedAt:   gameStart,
				Participant: p,
			})
		}
	}

	return results, nil
}

// GetResultsFromNames retrives results of Summoners attached to a Leaderboard.
func (s *Server) GetResultsFromNames(ctx context.Context, names []string, in *GetResultsArgs) (NameResults, error) {
	smnrs, err := s.API.GetSummoners(ctx, names)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve summoner PUUIDs")
	}

	var (
		nameResults NameResults = make(map[string][]Result)
		ids         []string
	)

	for _, smnr := range smnrs {
		nameResults[smnr.Name] = nil

		ids = append(ids, smnr.PUUID)
	}

	out, err := s.GetResults(ctx, ids, in)
	if err != nil {
		return nil, err
	}

	for _, smnr := range smnrs {
		nameResults[smnr.Name] = out[smnr.PUUID]
	}

	return nameResults, nil
}

// GetResultsFromLeaderboard retrieves results of Summoners attached to a Leaderboard.
func (s *Server) GetResultsFromLeaderboard(ctx context.Context, puuid string, in *GetResultsArgs) (NameResults, error) {
	board, err := s.Storage.GetLeaderboard(ctx, puuid)
	if err != nil {
		return nil, errors.Wrap(err, "get leaderboard failed")
	}

	var names []string
	for _, smnr := range board.Summoners {
		names = append(names, smnr.Name)
	}

	return s.GetResultsFromNames(ctx, names, in)
}

func (s *Server) GetSummoner(ctx context.Context, summonerName string) (*Summoner, error) {
	smnr, err := s.API.GetSummoner(ctx, summonerName)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get summoner")
	}

	le, err := s.API.GetLeagueEntry(ctx, smnr.ID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get league entry")
	}

	return &Summoner{
		Summoner:    *smnr,
		LeagueEntry: *le,
	}, nil
}
