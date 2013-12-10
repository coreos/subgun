package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/philips/go-mailgun"

	"github.com/philips/subgun/app"
)

var mg *mailgun.Client
var cfg *app.Config

func main() {
	// TODO: add a secret seed in here
	rand.Seed(time.Now().UTC().UnixNano())

	if len(os.Args) < 2 {
		fmt.Printf("Need config file argument")
		os.Exit(1)
	}

	cfg = app.ReadConfig(os.Args[1])
	mg = mailgun.New(cfg.Mailgun.Key)
	r := app.NewRouter(cfg, mg)

	port := cfg.ListenPort()
	http.ListenAndServe(":"+port, r)
}
