package http

import (
	"context"
	"encoding/json"
	"net/http"
	"shorty/internal/domain"
	"strconv"
)

const address string = "http://localhost"
const port string = ":8000"
const endpoint string = address + port

const shortenURL string = "/api/v1/shorten"

const resolveURL string = "/r"
const paramResolveURL string = resolveURL + "/{hash}"

const statURL string = "/api/v1/stats"
const paramStatURL string = statURL + "/{hash}"

type server struct {
	service domain.URLService
}

func NewServer(service domain.URLService) *server {
	return &server{
		service: service,
	}
}

func (c *server) Listen() {
	mux := http.NewServeMux()

	mux.HandleFunc("POST "+shortenURL, c.handlerShorten)
	mux.HandleFunc("GET "+paramResolveURL, c.handlerResolve)
	mux.HandleFunc("GET "+paramStatURL, c.handlerStat)

	http.ListenAndServe(port, mux)
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
		"short_url": endpoint + "/r/" + hash,
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
	if err != nil {
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
