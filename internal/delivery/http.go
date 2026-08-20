package delivery

import (
	"context"
	"encoding/json"
	"net/http"
	"shorty/internal/configuration"
	"shorty/internal/domain"
	"strconv"
)

type server struct {
	service domain.URLService
	config  *configuration.HttpConfig
}

func NewServer(service domain.URLService, config *configuration.HttpConfig) *server {
	return &server{
		service: service,
		config:  config,
	}
}

func (c *server) Listen() error {
	mux := http.NewServeMux()

	mux.HandleFunc("POST "+c.config.API.Shorten, c.handlerShorten)
	mux.HandleFunc("GET "+c.config.API.Resolve+"/{hash}", c.handlerResolve)
	mux.HandleFunc("GET "+c.config.API.Stat+"/{hash}", c.handlerStat)

	return http.ListenAndServe(":"+c.config.Port, mux)
}

func (c *server) handlerShorten(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	var req struct {
		LongURL string `json:"long_url"`
		TTLDays int    `json:"ttl_days"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	hash, err := c.service.Shorten(ctx, req.LongURL, req.TTLDays)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	resp := map[string]string{
		"short_url": c.config.Address + ":" + c.config.Port + "/r/" + hash,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func (c *server) handlerResolve(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	hash := r.PathValue("hash")

	url, err := c.service.Resolve(ctx, hash)
	if err != nil || url == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Location", url.LongURL)
	w.WriteHeader(http.StatusFound)
}

func (c *server) handlerStat(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	hash := r.PathValue("hash")

	url, err := c.service.GetURL(ctx, hash)
	if err != nil || url == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	resp := map[string]string{
		"long_url":        url.LongURL,
		"redirect_counts": strconv.FormatInt(url.Redirects, 10),
		"created_at":      url.CreatedAt.String(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)

}
