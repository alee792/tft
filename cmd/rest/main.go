package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/alee792/teamfit/internal/rest"
	"github.com/alee792/teamfit/pkg/leaderboards"
	"github.com/alee792/teamfit/pkg/tft"
	"gopkg.in/alecthomas/kingpin.v2"
)

func main() {
	var (
		restCfg rest.Config
		tftCfg  tft.Config

		app = kingpin.New("tft", "Test CLI for TFT API")
	)

	app.Flag("addr", "Address to serve from").Default(":8080").StringVar(&restCfg.Addr)
	app.Flag("key", "Riot API key").Envar("RIOT_API_KEY").Short('k').Required().StringVar(&tftCfg.APIKey)

	kingpin.MustParse(app.Parse(os.Args[1:]))

	// Create API Client and Leaderboard server.
	b := &leaderboards.Server{
		API:     tft.NewClient(http.DefaultClient, tftCfg),
		Storage: nil,
	}

	s, err := rest.NewServer(restCfg, rest.WithBoarder(b))
	if err != nil {
		panic(err)
	}

	fmt.Println(s.Config)

	if err := http.ListenAndServe(s.Config.Addr, s.Router); err != nil {
		s.Logger.Fatal(err)
	}
}
