package httpclient

import (
	"net/http"
	"time"
)

type APIClient struct {
	client  *http.Client
	baseURL string
}

// NewAPIClient creates a new API client
func NewAPIClient(baseURL string) *APIClient {
	return &APIClient{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		baseURL: baseURL,
	}
}
