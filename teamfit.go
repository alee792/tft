package teamfit

import (
	"context"
	"github.com/alee792/teamfit/pkg/tft"
)

type Server struct {
	api API
}

type API interface {
	GetSummoner(context.Context, *tft.GetSummonerRequest) (*tft.GetSummonerResponse, error)
}
