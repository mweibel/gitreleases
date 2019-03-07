package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	log "github.com/inconshreveable/log15"
)

const (
	githubGraphqlEndpoint = "https://api.github.com/graphql"
)

var (
	terminate = make(chan os.Signal, 1)
)

type gitreleasesEnv struct {
	addr            string
	token           string
	metricsUsername string
	metricsPassword string
}

func getEnv() gitreleasesEnv {
	addr := os.Getenv("LISTEN_ADDR")
	token := os.Getenv("GITHUB_TOKEN")
	metricsUsername := os.Getenv("METRICS_USERNAME")
	metricsPassword := os.Getenv("METRICS_PASSWORD")

	if addr == "" {
		panic("LISTEN_ADDR is required")
	}
	if token == "" {
		panic("GITHUB_TOKEN is required")
	}
	if metricsUsername == "" {
		panic("METRICS_USERNAME is required")
	}
	if metricsPassword == "" {
		panic("METRICS_PASSWORD is required")
	}
	return gitreleasesEnv{addr: addr, token: token, metricsUsername: metricsUsername, metricsPassword: metricsPassword}
}

func main() {
	logger := log.New("module", "gitrelases")
	// FIXME: change to another format before deployment
	handler := log.StreamHandler(os.Stdout, log.TerminalFormat())
	logger.SetHandler(handler)

	logger.Info("Starting up application")

	env := getEnv()

	httpClient := NewOauthClient(context.Background(), env.token)
	client := NewGitHubClient(githubGraphqlEndpoint, httpClient)
	apiServer := NewAPIServer(env.addr, env.metricsUsername, env.metricsPassword, client, logger)

	// Catch SIGINT and SIGTERM.
	signal.Notify(terminate, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		err := apiServer.Start()
		if err == http.ErrServerClosed {
			logger.Info("server closed")
		} else {
			logger.Error("cannot start server", "err", err)
			terminate <- syscall.SIGABRT
		}
	}()

	// Wait for SIGINT or SIGTERM as stop signal (or SIGABRT if the server
	// could not be started).
	sig := <-terminate
	if sig == syscall.SIGABRT {
		return
	}
	logger.Info("termination signal received", "signal", sig.String())

	// Graceful shutdown of the HTTP server. Give 500ms to finish
	// current messages being handled.
	grace, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	err := apiServer.Shutdown(grace)
	if err != nil {
		logger.Info("server shutdown with problems", "err", err)
	} else {
		logger.Info("server shut down properly", "err", err)
	}
}
