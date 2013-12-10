package app

import (
	"encoding/json"
	"io/ioutil"
	"net"
	"os"
)

type Config struct {
	Subscribegun struct {
		Listen string
		Lists  []string
	}
	Mailgun struct {
		Key string
	}
}

func (cfg *Config) ListenPort() string {
	_, port, _ := net.SplitHostPort(cfg.Subscribegun.Listen)
	if port == "" {
		port = "8080"
	}
	return port
}

func ReadConfig(path string) *Config {
	configBytes, err := ioutil.ReadFile(os.Args[1])

	if err != nil {
		panic(err)
	}

	var cfg Config
	err = json.Unmarshal(configBytes, &cfg)
	if err != nil {
		panic(err)
	}
	return &cfg
}
