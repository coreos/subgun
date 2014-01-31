package main

import (
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/philips/go-mailgun"

	"github.com/philips/subgun/app"
)

var mg *mailgun.Client
var cfg *app.Config

func main() {
	// TODO: add a secret seed in here
	rand.Seed(time.Now().UTC().UnixNano())

	cfg, err := app.GetConfigFromEnv(os.Environ())
	if err != nil {
		panic(err.Error())
	}

	mg = mailgun.New(cfg.Mailgun.Key)
	r := app.NewRouter(cfg, mg)

	if strings.HasPrefix(cfg.Subscribegun.Listen, "fd://") {
		app.ServeFD(r)
	} else {
		port := cfg.ListenPort()
		app.ServeTCP(r, port)
	}
}
