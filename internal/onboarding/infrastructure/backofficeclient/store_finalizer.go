package backofficeclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	onboardingconfig "github.com/tuannm99/podzone/internal/onboarding/config"
	storeentity "github.com/tuannm99/podzone/internal/onboarding/domain/store/entity"
	storeoutputport "github.com/tuannm99/podzone/internal/onboarding/domain/store/outputport"
)

type StoreFinalizer struct {
	baseURL string
	token   string
	client  *http.Client
}

var _ storeoutputport.OperationalStoreFinalizer = (*StoreFinalizer)(nil)

func NewStoreFinalizer(cfg onboardingconfig.AuthConfig) *StoreFinalizer {
	return &StoreFinalizer{
		baseURL: strings.TrimRight(cfg.BackofficeURL, "/"),
		token:   cfg.BackofficeServiceToken,
		client:  &http.Client{Timeout: 20 * time.Second},
	}
}

func (f *StoreFinalizer) FinalizeStore(ctx context.Context, request storeentity.StoreRequest) error {
	if f.baseURL == "" || f.token == "" {
		return fmt.Errorf("backoffice store finalizer is not configured")
	}
	payload, err := json.Marshal(map[string]string{
		"workspace_id": request.WorkspaceID,
		"store_id":     request.ID.Hex(),
		"name":         request.Name,
		"owner_id":     request.RequestedBy,
	})
	if err != nil {
		return fmt.Errorf("encode backoffice bootstrap request: %w", err)
	}

	httpRequest, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		f.baseURL+"/internal/backoffice/v1/stores:bootstrap",
		bytes.NewReader(payload),
	)
	if err != nil {
		return fmt.Errorf("create backoffice bootstrap request: %w", err)
	}
	httpRequest.Header.Set("Content-Type", "application/json")
	httpRequest.Header.Set("X-Backoffice-Service-Token", f.token)

	response, err := f.client.Do(httpRequest)
	if err != nil {
		return fmt.Errorf("call backoffice bootstrap: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode >= http.StatusOK && response.StatusCode < http.StatusMultipleChoices {
		return nil
	}
	body, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
	return fmt.Errorf(
		"backoffice bootstrap returned %s: %s",
		response.Status,
		strings.TrimSpace(string(body)),
	)
}
