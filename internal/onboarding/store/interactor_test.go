package store

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"

	storeentity "github.com/tuannm99/podzone/internal/onboarding/store/entity"
	storeinputport "github.com/tuannm99/podzone/internal/onboarding/store/inputport"
)

type fakeStoreRepo struct {
	requests map[string]storeentity.StoreRequest
}

func newFakeStoreRepo() *fakeStoreRepo {
	return &fakeStoreRepo{requests: map[string]storeentity.StoreRequest{}}
}

func (r *fakeStoreRepo) EnsureIndexes(context.Context) error { return nil }

func (r *fakeStoreRepo) FindBySubdomain(_ context.Context, subdomain string) (*storeentity.StoreRequest, error) {
	for _, req := range r.requests {
		if req.Subdomain == subdomain {
			copyReq := req
			return &copyReq, nil
		}
	}
	return nil, nil
}

func (r *fakeStoreRepo) Create(_ context.Context, request storeentity.StoreRequest) (*storeentity.StoreRequest, error) {
	request.ID = primitive.NewObjectID()
	r.requests[request.ID.Hex()] = request
	copyReq := request
	return &copyReq, nil
}

func (r *fakeStoreRepo) FindByID(_ context.Context, id string) (*storeentity.StoreRequest, error) {
	req, ok := r.requests[id]
	if !ok {
		return nil, nil
	}
	copyReq := req
	return &copyReq, nil
}

func (r *fakeStoreRepo) List(_ context.Context, workspaceID string) ([]storeentity.StoreRequest, error) {
	out := make([]storeentity.StoreRequest, 0, len(r.requests))
	for _, req := range r.requests {
		if workspaceID != "" && req.WorkspaceID != workspaceID {
			continue
		}
		out = append(out, req)
	}
	return out, nil
}

func (r *fakeStoreRepo) UpdateStatus(_ context.Context, id string, status storeentity.RequestStatus) error {
	req, ok := r.requests[id]
	if !ok {
		return nil
	}
	now := time.Now().UTC()
	req.Status = status
	req.UpdatedAt = now
	if status == storeentity.RequestStatusQueued {
		req.ApprovedAt = &now
	}
	if status == storeentity.RequestStatusReady || status == storeentity.RequestStatusFailed {
		req.CompletedAt = &now
	}
	r.requests[id] = req
	return nil
}

func setupStoreInteractor(t *testing.T) (*StoreInteractor, *fakeStoreRepo) {
	t.Helper()
	repo := newFakeStoreRepo()
	return &StoreInteractor{repo: repo}, repo
}

func TestCreateStoreRequest_ReturnsErrSubdomainTaken_WhenExistingRequestFound(t *testing.T) {
	svc, repo := setupStoreInteractor(t)

	existing := storeentity.StoreRequest{
		ID:          primitive.NewObjectID(),
		Name:        "Existing",
		Subdomain:   "taken",
		WorkspaceID: "workspace-1",
		RequestedBy: "user-1",
		Status:      storeentity.RequestStatusQueued,
	}
	repo.requests[existing.ID.Hex()] = existing

	_, err := svc.CreateStoreRequest(context.Background(), "New", "taken", "workspace-2", "user-2")
	require.ErrorIs(t, err, ErrSubdomainTaken)
}

func TestCreateStoreRequest_Success(t *testing.T) {
	svc, repo := setupStoreInteractor(t)

	request, err := svc.CreateStoreRequest(context.Background(), "New", "new", "workspace-1", "user-1")
	require.NoError(t, err)
	require.NotNil(t, request)
	require.NotEmpty(t, request.ID)
	require.Equal(t, storeinputport.RequestStatusRequested, request.Status)

	count := 0
	for _, req := range repo.requests {
		if req.Subdomain == "new" {
			count++
		}
	}
	require.Equal(t, 1, count)
}

func TestApproveStoreRequest_TransitionsToQueued(t *testing.T) {
	svc, repo := setupStoreInteractor(t)

	request := storeentity.StoreRequest{
		ID:          primitive.NewObjectID(),
		Name:        "Existing",
		Subdomain:   "queued",
		WorkspaceID: "workspace-1",
		RequestedBy: "user-1",
		Status:      storeentity.RequestStatusRequested,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	repo.requests[request.ID.Hex()] = request

	err := svc.ApproveStoreRequest(context.Background(), request.ID.Hex())
	require.NoError(t, err)

	updated, err := svc.GetStoreRequest(context.Background(), request.ID.Hex())
	require.NoError(t, err)
	require.Equal(t, storeinputport.RequestStatusQueued, updated.Status)
	require.NotNil(t, updated.ApprovedAt)
}
