package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"shorty/internal/domain"
	"testing"

	"github.com/stretchr/testify/require"
)

type mockService struct {
	getRunFunc     func(ctx context.Context)
	getShortenFunc func(ctx context.Context, longURL string, ttlDays int) (string, error)
	getResolveFunc func(ctx context.Context, hash string) (*domain.URL, error)
	getGetURLFunc  func(ctx context.Context, hash string) (*domain.URL, error)
	getGetLinkFunc func(ctx context.Context, hash string) (string, error)
}

func (m *mockService) Run(ctx context.Context) {
	m.getRunFunc(ctx)
}

func (m *mockService) Shorten(ctx context.Context, longURL string, ttlDays int) (string, error) {
	return m.getShortenFunc(ctx, longURL, ttlDays)
}

func (m *mockService) Resolve(ctx context.Context, hash string) (*domain.URL, error) {
	return m.getResolveFunc(ctx, hash)
}

func (m *mockService) GetURL(ctx context.Context, hash string) (*domain.URL, error) {
	return m.getGetURLFunc(ctx, hash)
}

func (m *mockService) GetLink(ctx context.Context, hash string) (string, error) {
	return m.getGetLinkFunc(ctx, hash)
}

func (m *mockService) Stop() {}

func listener(s *server) {
	s.Listen()

}

func prepareService() *mockService {
	return &mockService{
		getRunFunc: func(ctx context.Context) {},
		getShortenFunc: func(ctx context.Context, longURL string, ttlDays int) (string, error) {
			return longURL[:5], nil
		},
		getResolveFunc: func(ctx context.Context, hash string) (*domain.URL, error) {
			return &domain.URL{
				Hash:    hash,
				LongURL: hash + "_GOT_" + hash,
			}, nil
		},
		getGetURLFunc: func(ctx context.Context, hash string) (*domain.URL, error) {
			return &domain.URL{
				Hash:      hash,
				LongURL:   hash + "_GOT_" + hash,
				Redirects: int64(10),
			}, nil
		},
		getGetLinkFunc: func(ctx context.Context, hash string) (string, error) {
			return hash + "_GOT_" + hash, nil
		},
	}
}

func TestShorten_Http(t *testing.T) {
	serv := prepareService()

	longURL := "https://google.com"
	requestBody := []byte(`{"long_url": "` + longURL + `"}`)

	req, err := http.NewRequest(http.MethodPost, shortenURL, bytes.NewBuffer(requestBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	server := NewServer(serv)

	rr := httptest.NewRecorder()
	server.handlerShorten(rr, req)

	require.Equal(t, http.StatusCreated, rr.Code)

	var resp map[string]string
	err = json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)

	require.Equal(t, endpoint+"/r/"+longURL[:5], resp["short_url"])
}

func TestResolve_Http(t *testing.T) {
	serv := prepareService()

	hash := "1234567890"

	req, err := http.NewRequest(http.MethodGet, resolveURL+"/"+hash, nil)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	server := NewServer(serv)

	mux := http.NewServeMux()
	mux.HandleFunc(paramResolveURL, server.handlerResolve)

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	require.Equal(t, http.StatusFound, rr.Code)
	require.Equal(t, hash+"_GOT_"+hash, rr.Header().Get("Location"))
}

func TestStat_Http(t *testing.T) {
	serv := prepareService()

	hash := "1234567890"

	req, err := http.NewRequest(http.MethodGet, statURL+"/"+hash, nil)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	server := NewServer(serv)

	mux := http.NewServeMux()
	mux.HandleFunc(paramStatURL, server.handlerStat)

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	var resp map[string]string
	err = json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, rr.Code)
	require.Equal(t, hash+"_GOT_"+hash, resp["long_url"])
	require.Equal(t, "10", resp["redirect_counts"])
}
