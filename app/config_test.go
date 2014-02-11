package app

import (
	"testing"
)

func TestGetConfigFromEnv(t *testing.T) {
	fauxEnv := []string{
		"SUBGUN_LISTEN=127.0.0.1:8081",
		"SUBGUN_LISTS=foo@example.com,bar@example.com",
		"SUBGUN_API_KEY=secrete",
	}
	cfg, err := GetConfigFromEnv(fauxEnv)

	if err != nil {
		t.Fatalf("Received unexpected error: %v", err)
	}

	if cfg.Subscribegun.Listen != "127.0.0.1:8081" {
		t.Error("Subscribegun.Listen does not match expected value '127.0.0.1:8081'")
	}

	if len(cfg.Subscribegun.Lists) != 2 || cfg.Subscribegun.Lists[0] != "foo@example.com" || cfg.Subscribegun.Lists[1] != "bar@example.com" {
		t.Error("Subscribegun.Lists does not match expected value ['foo@example.com', 'bar@example.com']")
	}

	if cfg.Mailgun.Key != "secrete" {
		t.Error("Mailgun.Key does not match expected value 'secrete'")
	}
}

func TestGetConfigFromEnvMissingVariables(t *testing.T) {
	fauxEnv := []string{
		"SUBGUN_LISTEN=127.0.0.1:8081",
		"SUBGUN_LISTS=foo@example.com,bar@example.com",
		"SUBGUN_API_KEY=secrete",
	}

	// Iterate through each permutation of a single missing
	// env variable and assert an error is raised

	tmpEnv := fauxEnv[0:1]
	_, err := GetConfigFromEnv(tmpEnv)

	if err == nil {
		t.Fatalf("Expected error, received nil")
	}

	tmpEnv = fauxEnv[1:2]
	_, err = GetConfigFromEnv(tmpEnv)

	if err == nil {
		t.Fatalf("Expected error, received nil")
	}

	tmpEnv = append(fauxEnv[0:0], fauxEnv[2])
	_, err = GetConfigFromEnv(tmpEnv)

	if err == nil {
		t.Fatalf("Expected error, received nil")
	}
}

func TestGetConfigFromEnvListenPortDefault(t *testing.T) {
	fauxEnv := []string{
		"SUBGUN_LISTEN=127.0.0.1",
		"SUBGUN_LISTS=foo@example.com,bar@example.com",
		"SUBGUN_API_KEY=secrete",
	}
	cfg, _ := GetConfigFromEnv(fauxEnv)
	port := cfg.ListenPort()

	if port != "8080" {
		t.Fatalf("Expected ListenPort() to return 8080, got %s", port)
	}
}

func TestGetConfigFromEnvListenPortCustom(t *testing.T) {
	fauxEnv := []string{
		"SUBGUN_LISTEN=127.0.0.1:8081",
		"SUBGUN_LISTS=foo@example.com,bar@example.com",
		"SUBGUN_API_KEY=secrete",
	}
	cfg, _ := GetConfigFromEnv(fauxEnv)
	port := cfg.ListenPort()

	if port != "8081" {
		t.Fatalf("Expected ListenPort() to return 8081, got %s", port)
	}
}
