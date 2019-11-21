package leaderboards

import (
	"context"

	"github.com/alee792/teamfit/pkg/tft"
)

// Stats are aggregations of game results.
type Stats struct {
	Games             int
	DamageDealt       int
	PlayersEliminated int
	BoardValue        int
	Wins              int
	TopFours          int
	AverageFinish     float32
}

func getBoardValue(uu []tft.Unit) int {
	var val int
	for _, u := range uu {
		val += u.Tier * u.Rarity
	}
	return val
}

// GetStatsArgs is a request for aggregated results.
type GetStatsArgs struct {
	GameLimit int
}

// GetStatsResponse returns aggregated results.
type GetStatsResponse struct {
	Leaderboard map[string]Stats // Key = Summoner.Name
}

func (s *Server) GetStats(ctx context.Context, names []string, in *GetStatsArgs) (map[string]Stats, error) {
	out, err := s.GetResultsFromNames(ctx, names, &GetResultsArgs{
		GameLimit: in.GameLimit,
	})
	if err != nil {
		return nil, err
	}

	var stats = make(map[string]Stats)
	for n, rr := range out {
		stats[n] = CalculateStats(rr)
	}

	return stats, nil
}

// CalculateStats for a set of results.
func CalculateStats(rr []Result) Stats {
	var stat Stats
	var finishes float32

	for _, r := range rr {
		stat.DamageDealt += r.TotalDamageToPlayers
		stat.PlayersEliminated += r.PlayersEliminated
		stat.BoardValue += getBoardValue(r.Units)
		// Must divide by total games before returning!
		finishes += float32(r.Placement)

		switch place := r.Placement; {
		case place == 1:
			stat.TopFours++
			stat.Wins++
		case place > 5:
			stat.TopFours++
		}
	}

	stat.Games = len(rr)
	stat.AverageFinish = finishes / float32(len(rr))

	return stat
}
