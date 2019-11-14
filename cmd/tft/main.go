package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/alee792/teamfit/pkg/tft"
	"gopkg.in/alecthomas/kingpin.v2"
)

func main() {
	var (
		app      = kingpin.New("tft", "Test CLI for TFT API")
		summoner = app.Arg("summoner", "Summoner Name").Required().String()
		key      = app.Flag("key", "Riot API key").Envar("RIOT_API_KEY").String()
		verbose  = app.Flag("verbose", "show units and traits").Short('v').Bool()
	)

	app.HelpFlag.Short('h')
	kingpin.MustParse(app.Parse(os.Args[1:]))

	c := tft.NewClient(http.DefaultClient, tft.Config{
		APIKey: *key,
	})

	ctx := context.Background()
	sOut, err := c.GetSummoner(ctx, &tft.GetSummonerRequest{
		SummonerName: *summoner,
	})
	if err != nil {
		panic(err)
	}

	mmOut, err := c.ListMatches(ctx, &tft.ListMatchesRequest{
		PUUID: sOut.Summoner.PUUID,
	})
	if err != nil {
		panic(err)
	}

	mOut, err := c.GetMatch(ctx, &tft.GetMatchRequest{
		MatchID: mmOut.MatchIDs[0],
	})
	if err != nil {
		panic(err)
	}

	for _, p := range mOut.Match.Info.Participants {
		if p.PUUID == sOut.Summoner.PUUID {
			if !*verbose {
				p.Traits = nil
				p.Units = nil
			}
			p.PUUID = ""
			bb, err := json.MarshalIndent(p, "", "  ")
			if err != nil {
				panic(err)
			}
			fmt.Printf("%s's most recent game:\n%+v\n",
				sOut.Summoner.Name, string(bb),
			)
		}
	}
}
