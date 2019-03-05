package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/inconshreveable/log15"

	"github.com/gorilla/mux"
)

const requestTimeout = 1 * time.Second

type apiServer struct {
	server       *http.Server
	router       *mux.Router
	githubClient *GithubClient
	logger       log15.Logger
}

func (as *apiServer) Start() error {
	return as.server.ListenAndServe()
}
func (as *apiServer) Shutdown(ctx context.Context) error {
	return as.server.Shutdown(ctx)
}

func (as *apiServer) DownloadRelease(w http.ResponseWriter, r *http.Request) {
	reqLogger := as.logger.New("method", r.Method, "url", r.RequestURI)
	reqLogger.Info("fetching release URL")

	vars := mux.Vars(r)
	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()

	url, err := as.githubClient.FetchReleaseURL(ctx, vars["owner"], vars["repo"], vars["tag"], vars["assetName"])
	if err != nil {
		reqLogger.Error("error retrieving release URL", "err", err)
		w.WriteHeader(http.StatusBadGateway)
		fmt.Fprintln(w, "Unexpected error calling GitHub")
		return
	}

	reqLogger.Info("found release URL", "url", url)

	w.Header().Set("Location", url)
	w.WriteHeader(http.StatusMovedPermanently)
}

func NewAPIServer(addr string, client *GithubClient, logger log15.Logger) *apiServer {
	r := mux.NewRouter()

	as := apiServer{
		server: &http.Server{
			Addr:           addr,
			Handler:        r,
			ReadTimeout:    10 * time.Second,
			WriteTimeout:   10 * time.Second,
			MaxHeaderBytes: 1 << 20,
		},
		githubClient: client,
		logger:       logger,
	}

	r.HandleFunc("/d/{owner}/{repo}/{tag}/{assetName}", as.DownloadRelease)

	return &as
}
