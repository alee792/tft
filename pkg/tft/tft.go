package tft

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"path"
)

type Client struct {
	client *http.Client
	Config Config
}

type Config struct {
	APIKey string
}

func NewClient(client *http.Client, cfg Config) *Client {
	return &Client{
		client: client,
		Config: cfg,
	}
}

func (c *Client) GetSummoner(ctx context.Context, name string) (*Summoner, error) {
	r, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://na1.api.riotgames.com/tft/summoner/v1/summoners/by-name/", nil)
	if err != nil {
		return nil, err
	}

	r.Header.Set("X-Riot-Token", c.Config.APIKey)
	r.URL.Path = path.Join(r.URL.Path, name)

	resp, err := c.client.Do(r)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode > 299 || resp.StatusCode < 200 {
		return nil, fmt.Errorf("non 200 HTTP status code: %d", resp.StatusCode)
	}

	body := resp.Body
	defer resp.Body.Close()

	var s Summoner
	if err := json.NewDecoder(body).Decode(&s); err != nil {
		return nil, err
	}

	return &s, nil
}

func (c *Client) GetSummoners(ctx context.Context, names []string) ([]Summoner, error) {
	var summoners []Summoner
	for _, name := range names {
		out, err := c.GetSummoner(ctx, name)
		if err != nil {
			return nil, err
		}

		summoners = append(summoners, *out)
	}

	return summoners, nil
}

type ListMatchesRequest struct {
	PUUID string
}

type ListMatchesResponse struct {
	MatchIDs []string
}

func (c *Client) ListMatches(ctx context.Context, in *ListMatchesRequest) (*ListMatchesResponse, error) {
	if in.PUUID == "" {
		return nil, fmt.Errorf("invalid PUUID")
	}

	path := fmt.Sprintf("https://americas.api.riotgames.com/tft/match/v1/matches/by-puuid/%s/ids", in.PUUID)
	r, err := http.NewRequestWithContext(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	r.Header.Set("X-Riot-Token", c.Config.APIKey)

	resp, err := c.client.Do(r)
	if err != nil {
		return nil, err
	}

	body := resp.Body
	defer resp.Body.Close()

	if resp.StatusCode > 299 || resp.StatusCode < 200 {
		return nil, fmt.Errorf("non 200 HTTP status code: %d", resp.StatusCode)
	}

	var matches []string
	if err := json.NewDecoder(body).Decode(&matches); err != nil {
		return nil, err
	}

	return &ListMatchesResponse{
		MatchIDs: matches,
	}, nil
}

type GetMatchRequest struct {
	MatchID string
}

type GetMatchResponse struct {
	Match Match
}

func (c *Client) GetMatch(ctx context.Context, in *GetMatchRequest) (*GetMatchResponse, error) {
	if in.MatchID == "" {
		return nil, fmt.Errorf("invalid matchID")
	}

	path := fmt.Sprintf("https://americas.api.riotgames.com/tft/match/v1/matches/%s", in.MatchID)
	r, err := http.NewRequestWithContext(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	r.Header.Set("X-Riot-Token", c.Config.APIKey)

	resp, err := c.client.Do(r)
	if err != nil {
		return nil, err
	}

	body := resp.Body
	defer resp.Body.Close()

	var m Match
	if err := json.NewDecoder(body).Decode(&m); err != nil {
		return nil, err
	}

	return &GetMatchResponse{
		Match: m,
	}, nil
}

func (c *Client) GetMostRecentMatch(ctx context.Context, name string) (*Match, error) {
	smnr, err := c.GetSummoner(ctx, name)
	if err != nil {
		return nil, err
	}

	mmOut, err := c.ListMatches(ctx, &ListMatchesRequest{
		PUUID: smnr.PUUID,
	})
	if err != nil {
		return nil, err
	}

	mOut, err := c.GetMatch(ctx, &GetMatchRequest{
		MatchID: mmOut.MatchIDs[0],
	})
	if err != nil {
		return nil, err
	}

	return &mOut.Match, nil
}

// Summoner uses reflects the convention of the Riot API and
// uses camelCase instead of snake_case for JSON encoding.
type Summoner struct {
	ProileIconID  int    `json:"profileIconId"`
	Name          string `json:"name"`
	PUUID         string `json:"puuid"`
	SummonerLevel int    `json:"summonerLevel"`
	AccountID     string `json:"accountId"`
	ID            string `json:"id"`
	RevisionDate  int    `json:"revisionDate"`
}

// MarshalJSON hides confidential fields.
func (s *Summoner) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		ProileIconID  int    `json:"-"`
		Name          string `json:"name"`
		PUUID         string `json:"-"`
		SummonerLevel int    `json:"summonerLevel"`
		AccountID     string `json:"-"`
		ID            string `json:"-"`
		RevisionDate  int    `json:"revisionDate"`
	}{
		Name:          s.Name,
		SummonerLevel: s.SummonerLevel,
		RevisionDate:  s.RevisionDate,
	})
}

type Match struct {
	Info     Info     `json:"info"`
	Metadata Metadata `json:"metadata"`
}

type Info struct {
	GameTimestamp int           `json:"game_datetime"`
	Participants  []Participant `json:"participants"`
	Set           int           `json:"tft_set_number"`
	GameLength    float32       `json:"game_length"`
	QueueID       int           `json:"queue_id"`
	GameVersion   string        `json:"game_version"`
}

type Participant struct {
	Placement            int       `json:"placement"`
	Level                int       `json:"level"`
	LastRound            int       `json:"last_round"`
	TimeEliminated       float32   `json:"time_eliminated"`
	Companion            Companion `json:"companion"`
	Traits               []Trait   `json:"traits"`
	PlayersEliminated    int       `json:"players_eliminated"`
	PUUID                string    `json:"puuid"`
	TotalDamageToPlayers int       `json:"total_damage_to_players"`
	Units                []Unit    `json:"units"`
}

type Companion struct {
	SkinID    int    `json:"skin_id"`
	ContentID string `json:"content_id"`
	Species   string `json:"species"`
}

type Trait struct {
	TierTotal   int    `json:"tier_total"`
	Name        string `json:"name"`
	TierCurrent int    `json:"tier_current"`
	NumUnits    int    `json:"num_units"`
}

type Unit struct {
	Tier        int    `json:"tier"`
	Items       []int  `json:"items"`
	CharacterID string `json:"character_id"`
	Name        string `json:"name"`
	Rarity      int    `json:"rarity"`
}

type Metadata struct {
	DataVersion  string   `json:"data_version"`
	Participants []string `json:"participants"`
	MatchID      string   `json:"match_id"`
}
