package store

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"

	infrasentity "github.com/tuannm99/podzone/internal/onboarding/domain/infrasmanager/entity"
	infrasinputport "github.com/tuannm99/podzone/internal/onboarding/domain/infrasmanager/inputport"
	infrasmocks "github.com/tuannm99/podzone/internal/onboarding/domain/infrasmanager/inputport/mocks"
	storeentity "github.com/tuannm99/podzone/internal/onboarding/domain/store/entity"
	storeinputport "github.com/tuannm99/podzone/internal/onboarding/domain/store/inputport"
	storemocks "github.com/tuannm99/podzone/internal/onboarding/domain/store/outputport/mocks"
)

func setupStoreInteractor(t *testing.T) (*StoreInteractor, *storemocks.MockStoreRepository) {
	t.Helper()
	repo := storemocks.NewMockStoreRepository(t)
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
	repo.EXPECT().FindBySubdomain(mock.Anything, "taken").Return(&existing, nil)

	_, err := svc.CreateStoreRequest(context.Background(), "New", "taken", "workspace-2", "user-2")
	require.ErrorIs(t, err, ErrSubdomainTaken)
}

func TestCreateStoreRequest_Success(t *testing.T) {
	svc, repo := setupStoreInteractor(t)

	repo.EXPECT().FindBySubdomain(mock.Anything, "new").Return(nil, nil)
	repo.EXPECT().
		Create(mock.Anything, mock.MatchedBy(func(request storeentity.StoreRequest) bool {
			return request.Name == "New" &&
				request.Subdomain == "new" &&
				request.WorkspaceID == "workspace-1" &&
				request.RequestedBy == "user-1" &&
				request.Status == storeentity.RequestStatusRequested
		})).
		RunAndReturn(func(_ context.Context, request storeentity.StoreRequest) (*storeentity.StoreRequest, error) {
			request.ID = primitive.NewObjectID()
			return &request, nil
		})

	request, err := svc.CreateStoreRequest(context.Background(), "New", "new", "workspace-1", "user-1")
	require.NoError(t, err)
	require.NotNil(t, request)
	require.NotEmpty(t, request.ID)
	require.Equal(t, storeinputport.RequestStatusRequested, request.Status)
}

func TestApproveStoreRequest_TransitionsToQueued(t *testing.T) {
	svc, repo := setupStoreInteractor(t)

	now := time.Now().UTC()
	request := storeentity.StoreRequest{
		ID:          primitive.NewObjectID(),
		Name:        "Existing",
		Subdomain:   "queued",
		WorkspaceID: "workspace-1",
		RequestedBy: "user-1",
		Status:      storeentity.RequestStatusRequested,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	queued := request
	queued.Status = storeentity.RequestStatusQueued
	queued.ApprovedAt = &now

	repo.EXPECT().FindByID(mock.Anything, request.ID.Hex()).Return(&request, nil).Once()
	repo.EXPECT().UpdateStatus(mock.Anything, request.ID.Hex(), storeentity.RequestStatusQueued).Return(nil)
	repo.EXPECT().FindByID(mock.Anything, request.ID.Hex()).Return(&queued, nil).Once()

	err := svc.ApproveStoreRequest(context.Background(), request.ID.Hex())
	require.NoError(t, err)

	updated, err := svc.GetStoreRequest(context.Background(), request.ID.Hex())
	require.NoError(t, err)
	require.Equal(t, storeinputport.RequestStatusQueued, updated.Status)
	require.NotNil(t, updated.ApprovedAt)
}

func TestProcessNextStoreRequest_ProvisionsQueuedRequest(t *testing.T) {
	repo := storemocks.NewMockStoreRepository(t)
	infra := infrasmocks.NewMockUsecase(t)
	svc := &StoreInteractor{
		repo:  repo,
		infra: infra,
		provisioner: storeentity.ProvisioningConfig{
			Enabled:      true,
			ClusterName:  "pg-default",
			Mode:         "schema",
			DBName:       "postgres",
			SchemaPrefix: "t_",
			Endpoint:     "postgres://postgres:***@pgbouncer:6432/postgres",
			SecretRef:    "postgres/default",
		},
	}

	id := primitive.NewObjectID()
	request := &storeentity.StoreRequest{
		ID:          id,
		Name:        "Urban Finds",
		Subdomain:   "urban-finds",
		WorkspaceID: "2e0df8f6-4964-447d-a287-67eabd0e65c9",
		RequestedBy: "user-1",
		Status:      storeentity.RequestStatusProvisioning,
	}
	ready := *request
	ready.Status = storeentity.RequestStatusReady
	ready.StoreID = &id

	repo.EXPECT().ClaimNextQueued(mock.Anything).Return(request, nil)
	infra.EXPECT().
		ManualUpsertConnection(
			mock.Anything,
			request.WorkspaceID,
			mock.MatchedBy(func(req infrasinputport.UpsertConnectionRequest) bool {
				return req.InfraType == infrasentity.InfraPostgres &&
					req.Name == "default" &&
					req.Endpoint == "postgres://postgres:***@pgbouncer:6432/postgres" &&
					req.SecretRef == "postgres/default" &&
					req.ClusterName == "pg-default" &&
					req.Mode == "schema" &&
					req.DBName == "postgres" &&
					req.SchemaName == "t_tenant_2e0df8f6_4964_447d_a287_67eabd0e65c9" &&
					req.Meta["store_request_id"] == id.Hex() &&
					req.Meta["store_id"] == id.Hex() &&
					req.Meta["store_subdomain"] == "urban-finds"
			}),
			mock.MatchedBy(func(actor map[string]string) bool {
				return actor["service"] == "onboarding" && actor["worker"] == "store-provisioner"
			}),
		).
		Return(&infrasinputport.UpsertConnectionResponse{CorrelationID: "correlation-1", Queued: true}, nil)
	repo.EXPECT().MarkReady(mock.Anything, id.Hex(), id.Hex()).Return(nil)
	repo.EXPECT().FindByID(mock.Anything, id.Hex()).Return(&ready, nil)

	processed, err := svc.ProcessNextStoreRequest(context.Background())
	require.NoError(t, err)
	require.NotNil(t, processed)
	require.Equal(t, storeinputport.RequestStatusReady, processed.Status)
	require.Equal(t, id.Hex(), processed.StoreID)
}
