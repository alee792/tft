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

		results     = app.Command("results", "fetches recent match results").Default()
		resultsArgs = setupCommonArgs(results)

		stats     = app.Command("stats", "calculate stats from recent matches")
		statsArgs = setupCommonArgs(stats)

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

	// Setup writer to stdout.
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")

	switch cmd {
	case results.FullCommand():
		out, err := boarder.GetResultsFromNames(ctx, resultsArgs.Names, &leaderboards.GetResultsArgs{
			GameLimit: resultsArgs.Matches,
		})
		if err != nil {
			panic(err)
		}

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
	case stats.FullCommand():
		out, err := boarder.GetStats(ctx, statsArgs.Names, &leaderboards.GetStatsArgs{
			GameLimit: statsArgs.Matches,
		})
		if err != nil {
			panic(err)
		}

		if err := enc.Encode(&out); err != nil {
			panic(err)
		}
	}
}

// CommonArgs shared between commands.
type CommonArgs struct {
	Matches int
	Names   []string
}

func setupCommonArgs(cmd *kingpin.CmdClause) *CommonArgs {
	var args CommonArgs
	cmd.Arg("smnrs", "summoner names").StringsVar(&args.Names)
	cmd.Flag("matches", "number of matches to pull").Default("1").Short('m').IntVar(&args.Matches)

	return &args
}
