package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/alee792/teamfit/pkg/leaderboards"
	"github.com/alee792/teamfit/pkg/tft"
	"gopkg.in/alecthomas/kingpin.v2"
)

func main() {
	var (
		app     = kingpin.New("tft", "Test CLI for TFT API")
		key     = app.Flag("key", "Riot API key").Envar("RIOT_API_KEY").Short('k').Required().String()
		verbose = app.Flag("verbose", "show units and traits").Short('v').Bool()
		_       = app.HelpFlag.Short('h')

		results = app.Command("results", "fetches the X most recent matches").Default()
		smnrs   = results.Arg("smnrs", "summoner names").Strings()
		matches = results.Flag("matches", "number of matches to pull").Default("1").Short('m').Int()

		// summoner = app.Arg("summoner", "Summoner Name").Required().String()

		ctx = context.Background()
	)

	// Parse flags.
	cmd := kingpin.MustParse(app.Parse(os.Args[1:]))

	// Create API Client and Leaderboard server.
	boarder := leaderboards.Server{
		API: tft.NewClient(http.DefaultClient, tft.Config{
			APIKey: *key,
		}),
		Storage: nil,
	}

	switch cmd {
	case results.FullCommand():
		out, err := boarder.GetResultsFromNames(ctx, *smnrs, &leaderboards.GetResultsArgs{
			GameLimit: *matches,
		})
		if err != nil {
			panic(err)
		}

		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")

		for name, results := range out {
			fmt.Printf("%s's Results\n", name)

			// Clean the results.
			for i := range results {
				results[i].PUUID = ""
				if !*verbose {
					results[i].Traits = nil
					results[i].Units = nil
				}
			}

			if err := enc.Encode(&results); err != nil {
				panic(err)
			}
		}

		if !*verbose {
			fmt.Printf("(Use -v to see units and traits)\n\n")
		}
		// default:

		// 	sOut, err := c.GetSummoner(ctx, &tft.GetSummonerRequest{
		// 		SummonerName: *summoner,
		// 	})
		// 	if err != nil {
		// 		panic(err)
		// 	}

		// 	mmOut, err := c.ListMatches(ctx, &tft.ListMatchesRequest{
		// 		PUUID: sOut.Summoner.PUUID,
		// 	})
		// 	if err != nil {
		// 		panic(err)
		// 	}

		// 	mOut, err := c.GetMatch(ctx, &tft.GetMatchRequest{
		// 		MatchID: mmOut.MatchIDs[0],
		// 	})
		// 	if err != nil {
		// 		panic(err)
		// 	}

		// 	var p tft.Participant
		// 	for _, v := range mOut.Match.Info.Participants {
		// 		if v.PUUID == sOut.Summoner.PUUID {
		// 			p = v
		// 			break
		// 		}
		// 	}

		// 	if !*verbose {
		// 		fmt.Printf("(Use -v to see units and traits)\n\n")

		// 		p.Traits = nil
		// 		p.Units = nil
		// 	}

		// 	p.PUUID = ""

		// 	bb, err := json.MarshalIndent(p, "", "  ")
		// 	if err != nil {
		// 		panic(err)
		// 	}

		// 	fmt.Printf("%s's Most Recent Game\n%s\n%+v\n",
		// 		sOut.Summoner.Name,
		// 		time.Unix(int64(mOut.Match.Info.GameTimestamp/1000), 0).Format(time.RFC1123),
		// 		string(bb),
		// 	)
	}

}
