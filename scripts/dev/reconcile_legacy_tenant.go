//go:build ignore

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type config struct {
	TenantID     string
	OwnerID      string
	StoreName    string
	Subdomain    string
	Actor        string
	ServiceURL   string
	ServiceToken string
	Timeout      time.Duration
}

type storeRequest struct {
	ID          string `json:"id"`
	WorkspaceID string `json:"workspace_id"`
	Name        string `json:"name"`
	Subdomain   string `json:"subdomain"`
	RequestedBy string `json:"requested_by"`
	OwnerID     string `json:"owner_id"`
	Status      string `json:"status"`
	StoreID     string `json:"store_id"`
	LastError   string `json:"last_error"`
}

type listStoreRequestsResponse struct {
	Items []storeRequest `json:"items"`
}

func main() {
	cfg, err := loadConfig()
	if err != nil {
		fail(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	request, err := createOrFindRequest(ctx, cfg)
	if err != nil {
		fail(err)
	}
	if request.Status == "ready" {
		printReady(request)
		return
	}

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			fail(fmt.Errorf("timed out waiting for tenant %q placement: %w", cfg.TenantID, ctx.Err()))
		case <-ticker.C:
			request, err = findRequest(ctx, cfg)
			if err != nil {
				fmt.Fprintf(os.Stderr, "reconcile poll failed, retrying: %v\n", err)
				continue
			}
			switch {
			case request.Status == "ready":
				printReady(request)
				return
			case isTerminalFailure(request.Status):
				fail(fmt.Errorf(
					"onboarding request %s reached %s: %s",
					request.ID,
					request.Status,
					request.LastError,
				))
			}
		}
	}
}

func loadConfig() (config, error) {
	timeout := 120 * time.Second
	if raw := strings.TrimSpace(os.Getenv("WAIT_SECONDS")); raw != "" {
		parsed, err := time.ParseDuration(raw + "s")
		if err != nil {
			return config{}, fmt.Errorf("parse WAIT_SECONDS: %w", err)
		}
		timeout = parsed
	}

	cfg := config{
		TenantID:     strings.TrimSpace(os.Getenv("TENANT_ID")),
		OwnerID:      strings.TrimSpace(os.Getenv("OWNER_ID")),
		StoreName:    strings.TrimSpace(os.Getenv("STORE_NAME")),
		Subdomain:    strings.TrimSpace(os.Getenv("STORE_SUBDOMAIN")),
		Actor:        envOr("RECONCILE_ACTOR", "legacy-placement-migration"),
		ServiceURL:   strings.TrimRight(envOr("ONBOARDING_URL", "http://localhost:8800"), "/"),
		ServiceToken: envOr("ONBOARDING_SERVICE_TOKEN", "dev-bootstrap-token"),
		Timeout:      timeout,
	}
	switch {
	case cfg.TenantID == "":
		return config{}, errors.New("TENANT_ID is required")
	case cfg.OwnerID == "":
		return config{}, errors.New("OWNER_ID is required")
	case cfg.StoreName == "":
		return config{}, errors.New("STORE_NAME is required")
	case cfg.Subdomain == "":
		return config{}, errors.New("STORE_SUBDOMAIN is required")
	}
	return cfg, nil
}

func createOrFindRequest(ctx context.Context, cfg config) (storeRequest, error) {
	payload, err := json.Marshal(map[string]string{
		"name":      cfg.StoreName,
		"subdomain": cfg.Subdomain,
		"owner_id":  cfg.OwnerID,
	})
	if err != nil {
		return storeRequest{}, fmt.Errorf("encode create request: %w", err)
	}

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		cfg.ServiceURL+"/onboarding/v1/stores",
		bytes.NewReader(payload),
	)
	if err != nil {
		return storeRequest{}, fmt.Errorf("create onboarding request: %w", err)
	}
	setHeaders(request, cfg)

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return storeRequest{}, fmt.Errorf("submit onboarding request: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode == http.StatusCreated {
		var created storeRequest
		if err := json.NewDecoder(response.Body).Decode(&created); err != nil {
			return storeRequest{}, fmt.Errorf("decode created onboarding request: %w", err)
		}
		fmt.Printf("Created onboarding request %s for tenant %s.\n", created.ID, cfg.TenantID)
		return created, nil
	}
	if response.StatusCode == http.StatusConflict {
		existing, findErr := findRequest(ctx, cfg)
		if findErr != nil {
			return storeRequest{}, fmt.Errorf("resolve conflicting onboarding request: %w", findErr)
		}
		fmt.Printf("Using existing onboarding request %s for tenant %s.\n", existing.ID, cfg.TenantID)
		return existing, nil
	}
	body, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
	return storeRequest{}, fmt.Errorf(
		"create onboarding request returned %s: %s",
		response.Status,
		strings.TrimSpace(string(body)),
	)
}

func findRequest(ctx context.Context, cfg config) (storeRequest, error) {
	query := url.Values{
		"collection.page":     {"1"},
		"collection.pageSize": {"100"},
	}
	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		cfg.ServiceURL+"/onboarding/v1/requests?"+query.Encode(),
		nil,
	)
	if err != nil {
		return storeRequest{}, fmt.Errorf("create list request: %w", err)
	}
	setHeaders(request, cfg)

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return storeRequest{}, fmt.Errorf("list onboarding requests: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
		return storeRequest{}, fmt.Errorf(
			"list onboarding requests returned %s: %s",
			response.Status,
			strings.TrimSpace(string(body)),
		)
	}

	var page listStoreRequestsResponse
	if err := json.NewDecoder(response.Body).Decode(&page); err != nil {
		return storeRequest{}, fmt.Errorf("decode onboarding requests: %w", err)
	}
	for _, item := range page.Items {
		if item.WorkspaceID == cfg.TenantID && item.Subdomain == cfg.Subdomain {
			return item, nil
		}
	}
	return storeRequest{}, fmt.Errorf(
		"tenant %q has no onboarding request for subdomain %q",
		cfg.TenantID,
		cfg.Subdomain,
	)
}

func setHeaders(request *http.Request, cfg config) {
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Onboarding-Service-Token", cfg.ServiceToken)
	request.Header.Set("X-Tenant-ID", cfg.TenantID)
	request.Header.Set("X-User-ID", cfg.Actor)
}

func isTerminalFailure(status string) bool {
	return strings.HasPrefix(status, "failed") ||
		status == "rejected" ||
		status == "cancelled" ||
		status == "archived"
}

func printReady(request storeRequest) {
	fmt.Printf(
		"Tenant placement ready: tenant=%s store=%s owner=%s request=%s\n",
		request.WorkspaceID,
		request.StoreID,
		request.OwnerID,
		request.ID,
	)
}

func envOr(key string, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
