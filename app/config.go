package app

import (
	"errors"
	"net"
	"strings"
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

func GetConfigFromEnv(env []string) (*Config, error) {
	var cfg Config

	vars := make(map[string]string, 3)
	for _, e := range env {
		tokens := strings.SplitN(e, "=", 2)
		vars[tokens[0]] = tokens[1]
	}

	listen, ok := vars["SUBGUN_LISTEN"]
	if !ok {
		return nil, errors.New("Environment variable SUBGUN_LISTEN could not be found")
	}
	cfg.Subscribegun.Listen = listen

	listCSV, ok := vars["SUBGUN_LISTS"]
	if !ok {
		return nil, errors.New("Environment variable SUBGUN_LISTS could not be found")
	}
	lists := strings.Split(listCSV, ",")
	if len(lists) == 0 {
		return nil, errors.New("Environment variable SUBGUN_LISTS provides no lists")
	}
	cfg.Subscribegun.Lists = lists

	key, ok := vars["SUBGUN_API_KEY"]
	if !ok {
		return nil, errors.New("Environment variable SUBGUN_API_KEY could not be found")
	}
	cfg.Mailgun.Key = key

	return &cfg, nil
}
