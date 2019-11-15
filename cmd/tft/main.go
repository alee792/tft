package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/alee792/teamfit/pkg/tft"
	"gopkg.in/alecthomas/kingpin.v2"
)

func main() {
	var (
		app      = kingpin.New("tft", "Test CLI for TFT API")
		summoner = app.Arg("summoner", "Summoner Name").Required().String()
		key      = app.Flag("key", "Riot API key").Envar("RIOT_API_KEY").Required().String()
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

	var p tft.Participant
	for _, v := range mOut.Match.Info.Participants {
		if v.PUUID == sOut.Summoner.PUUID {
			p = v
			break
		}
	}

	if !*verbose {
		fmt.Printf("(Use -v to see units and traits)\n\n")

		p.Traits = nil
		p.Units = nil
	}

	p.PUUID = ""

	bb, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		panic(err)
	}

	fmt.Printf("%s's Most Recent Game\n%s\n%+v\n",
		sOut.Summoner.Name,
		time.Unix(int64(mOut.Match.Info.GameTimestamp/1000), 0).Format(time.RFC1123),
		string(bb),
	)

}
