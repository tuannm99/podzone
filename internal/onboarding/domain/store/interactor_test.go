package store

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"

	infrasinputport "github.com/tuannm99/podzone/internal/onboarding/domain/infrasmanager/inputport"
	infrasmocks "github.com/tuannm99/podzone/internal/onboarding/domain/infrasmanager/inputport/mocks"
	storeentity "github.com/tuannm99/podzone/internal/onboarding/domain/store/entity"
	storeinputport "github.com/tuannm99/podzone/internal/onboarding/domain/store/inputport"
	storemocks "github.com/tuannm99/podzone/internal/onboarding/domain/store/outputport/mocks"
	"github.com/tuannm99/podzone/pkg/toolkit"
)

func setupStoreInteractor(t *testing.T) (*StoreInteractor, *storemocks.MockStoreRepository) {
	t.Helper()
	repo := storemocks.NewMockStoreRepository(t)
	allowTransitionLog(repo)
	return &StoreInteractor{repo: repo}, repo
}

func allowTransitionLog(repo *storemocks.MockStoreRepository) {
	repo.EXPECT().RecordTransition(mock.Anything, mock.Anything).Return(nil).Maybe()
}

func authenticatedContext(workspaceID string, userID string) context.Context {
	ctx := toolkit.WithTenantID(context.Background(), workspaceID)
	return toolkit.WithUserID(ctx, userID)
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

	_, err := svc.CreateStoreRequest(
		authenticatedContext("workspace-2", "2"),
		storeinputport.CreateStoreRequestCommand{Name: "New", Subdomain: "taken"},
	)
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
				request.RequestedBy == "1" &&
				request.Status == storeentity.RequestStatusRequested
		})).
		RunAndReturn(func(_ context.Context, request storeentity.StoreRequest) (*storeentity.StoreRequest, error) {
			request.ID = primitive.NewObjectID()
			return &request, nil
		})

	request, err := svc.CreateStoreRequest(
		authenticatedContext("workspace-1", "1"),
		storeinputport.CreateStoreRequestCommand{Name: "New", Subdomain: "new"},
	)
	require.NoError(t, err)
	require.NotNil(t, request)
	require.NotEmpty(t, request.ID)
	require.Equal(t, storeinputport.RequestStatusRequested, request.Status)
}

func TestCreateStoreRequest_AuthorizesAuthenticatedWorkspace(t *testing.T) {
	repo := storemocks.NewMockStoreRepository(t)
	allowTransitionLog(repo)
	authorizer := storemocks.NewMockAccessAuthorizer(t)
	svc := &StoreInteractor{repo: repo, authorizer: authorizer}
	ctx := authenticatedContext("workspace-1", "7")

	authorizer.EXPECT().
		AuthorizeStoreRequest(mock.Anything, "workspace-1", "7").
		Return(context.Canceled)

	_, err := svc.CreateStoreRequest(
		ctx,
		storeinputport.CreateStoreRequestCommand{Name: "New", Subdomain: "new"},
	)
	require.ErrorIs(t, err, context.Canceled)
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

	err := svc.ApproveStoreRequest(toolkit.WithUserID(context.Background(), "1"), request.ID.Hex())
	require.NoError(t, err)

	updated, err := svc.GetStoreRequest(
		authenticatedContext(request.WorkspaceID, request.RequestedBy),
		request.ID.Hex(),
	)
	require.NoError(t, err)
	require.Equal(t, storeinputport.RequestStatusQueued, updated.Status)
	require.NotNil(t, updated.ApprovedAt)
}

func TestRetryStoreRequest_TransitionsOwnedFailedRequestToQueued(t *testing.T) {
	repo := storemocks.NewMockStoreRepository(t)
	allowTransitionLog(repo)
	authorizer := storemocks.NewMockAccessAuthorizer(t)
	svc := &StoreInteractor{repo: repo, authorizer: authorizer}
	ctx := authenticatedContext("workspace-1", "7")
	request := storeentity.StoreRequest{
		ID:          primitive.NewObjectID(),
		WorkspaceID: "workspace-1",
		RequestedBy: "7",
		Status:      storeentity.RequestStatusFailed,
	}

	repo.EXPECT().FindByID(mock.Anything, request.ID.Hex()).Return(&request, nil)
	authorizer.EXPECT().
		AuthorizeStoreRequest(mock.Anything, "workspace-1", "7").
		Return(nil)
	repo.EXPECT().
		UpdateStatus(mock.Anything, request.ID.Hex(), storeentity.RequestStatusQueued).
		Return(nil)

	require.NoError(t, svc.RetryStoreRequest(ctx, request.ID.Hex()))
}

func TestRetryStoreRequest_HidesRequestOwnedByAnotherWorkspace(t *testing.T) {
	repo := storemocks.NewMockStoreRepository(t)
	allowTransitionLog(repo)
	svc := &StoreInteractor{repo: repo}
	request := storeentity.StoreRequest{
		ID:          primitive.NewObjectID(),
		WorkspaceID: "workspace-2",
		Status:      storeentity.RequestStatusFailed,
	}

	repo.EXPECT().FindByID(mock.Anything, request.ID.Hex()).Return(&request, nil)

	err := svc.RetryStoreRequest(
		authenticatedContext("workspace-1", "7"),
		request.ID.Hex(),
	)
	require.ErrorIs(t, err, ErrStoreNotFound)
}

func TestRetryStoreRequest_RejectsNonFailedRequest(t *testing.T) {
	repo := storemocks.NewMockStoreRepository(t)
	allowTransitionLog(repo)
	authorizer := storemocks.NewMockAccessAuthorizer(t)
	svc := &StoreInteractor{repo: repo, authorizer: authorizer}
	ctx := authenticatedContext("workspace-1", "7")
	request := storeentity.StoreRequest{
		ID:          primitive.NewObjectID(),
		WorkspaceID: "workspace-1",
		Status:      storeentity.RequestStatusProvisioning,
	}

	repo.EXPECT().FindByID(mock.Anything, request.ID.Hex()).Return(&request, nil)
	authorizer.EXPECT().
		AuthorizeStoreRequest(mock.Anything, "workspace-1", "7").
		Return(nil)

	err := svc.RetryStoreRequest(ctx, request.ID.Hex())
	require.ErrorIs(t, err, ErrInvalidStatus)
}

func TestProcessNextStoreRequest_ProvisionsQueuedRequest(t *testing.T) {
	repo := storemocks.NewMockStoreRepository(t)
	allowTransitionLog(repo)
	infra := infrasmocks.NewMockUsecase(t)
	svc := &StoreInteractor{
		repo:  repo,
		infra: infra,
		provisioner: storeentity.ProvisioningConfig{
			Enabled:      true,
			ClusterName:  "pg-default",
			Mode:         "schema",
			DBName:       "podzone_tenants",
			SchemaPrefix: "t_",
		},
	}

	id := primitive.NewObjectID()
	request := &storeentity.StoreRequest{
		ID:          id,
		Name:        "Urban Finds",
		Subdomain:   "urban-finds",
		WorkspaceID: "2e0df8f6-4964-447d-a287-67eabd0e65c9",
		RequestedBy: "user-1",
		Status:      storeentity.RequestStatusPlanning,
	}
	repo.EXPECT().ClaimNextQueued(mock.Anything).Return(request, nil)
	infra.EXPECT().
		ProvisionStorePlacement(
			mock.Anything,
			mock.MatchedBy(func(req infrasinputport.ProvisionStorePlacementRequest) bool {
				return req.RequestID == id.Hex() &&
					req.TenantID == request.WorkspaceID &&
					req.StoreID == id.Hex() &&
					req.Subdomain == "urban-finds" &&
					req.RequestedBy == "user-1"
			}),
			mock.MatchedBy(func(actor map[string]string) bool {
				return actor["service"] == "onboarding" && actor["worker"] == "store-provisioner"
			}),
		).
		Return(&infrasinputport.ProvisionStorePlacementResponse{
			CorrelationID: "correlation-1",
			AllocationID:  "allocation-1",
			Status:        "ready",
			Queued:        true,
		}, nil)
	repo.EXPECT().UpdateStatus(mock.Anything, id.Hex(), storeentity.RequestStatusPlanned).Return(nil)
	repo.EXPECT().UpdateStatus(mock.Anything, id.Hex(), storeentity.RequestStatusProvisioning).Return(nil)
	processed, err := svc.ProcessNextStoreRequest(context.Background())
	require.NoError(t, err)
	require.NotNil(t, processed)
	require.Equal(t, storeinputport.RequestStatusProvisioning, processed.Status)
	require.Empty(t, processed.StoreID)
}

func TestProcessNextStoreRequest_MarksPlatformSetupWhenProviderUnavailable(t *testing.T) {
	repo := storemocks.NewMockStoreRepository(t)
	allowTransitionLog(repo)
	infra := infrasmocks.NewMockUsecase(t)
	svc := &StoreInteractor{
		repo:  repo,
		infra: infra,
		provisioner: storeentity.ProvisioningConfig{
			Enabled: true,
		},
	}

	id := primitive.NewObjectID()
	request := &storeentity.StoreRequest{
		ID:          id,
		Name:        "Urban Finds",
		Subdomain:   "urban-finds",
		WorkspaceID: "workspace-1",
		RequestedBy: "user-1",
		Status:      storeentity.RequestStatusPlanning,
	}
	providerErr := errors.New("kubernetes placement provider is declared but not implemented")

	repo.EXPECT().ClaimNextQueued(mock.Anything).Return(request, nil)
	infra.EXPECT().
		ProvisionStorePlacement(mock.Anything, mock.Anything, mock.Anything).
		Return(nil, providerErr)
	repo.EXPECT().
		MarkBlocked(
			mock.Anything,
			id.Hex(),
			storeentity.RequestStatusPendingPlatformSetup,
			providerErr.Error(),
		).
		Return(nil)

	processed, err := svc.ProcessNextStoreRequest(context.Background())

	require.ErrorIs(t, err, providerErr)
	require.Nil(t, processed)
}

func TestFinalizeNextStoreRequest_BootstrapsOperationalStoreBeforeReady(t *testing.T) {
	repo := storemocks.NewMockStoreRepository(t)
	allowTransitionLog(repo)
	infra := infrasmocks.NewMockUsecase(t)
	finalizer := storemocks.NewMockOperationalStoreFinalizer(t)
	svc := &StoreInteractor{
		repo:      repo,
		infra:     infra,
		finalizer: finalizer,
	}

	id := primitive.NewObjectID()
	request := &storeentity.StoreRequest{
		ID:          id,
		Name:        "Urban Finds",
		Subdomain:   "urban-finds",
		WorkspaceID: "workspace-1",
		RequestedBy: "user-1",
		Status:      storeentity.RequestStatusProvisioning,
	}
	repo.EXPECT().FindNextProvisioning(mock.Anything).Return(request, nil)
	infra.EXPECT().EnsurePlacementRoute(mock.Anything, "workspace-1", id.Hex()).Return(true, nil)
	finalizer.EXPECT().FinalizeStore(mock.Anything, *request).Return(nil)
	repo.EXPECT().MarkReady(mock.Anything, id.Hex(), id.Hex()).Return(nil)

	finalized, err := svc.FinalizeNextStoreRequest(context.Background())
	require.NoError(t, err)
	require.NotNil(t, finalized)
	require.Equal(t, storeinputport.RequestStatusReady, finalized.Status)
	require.Equal(t, id.Hex(), finalized.StoreID)
}

func TestFinalizeNextStoreRequest_WaitsForPlacementRoute(t *testing.T) {
	repo := storemocks.NewMockStoreRepository(t)
	allowTransitionLog(repo)
	infra := infrasmocks.NewMockUsecase(t)
	finalizer := storemocks.NewMockOperationalStoreFinalizer(t)
	svc := &StoreInteractor{
		repo:      repo,
		infra:     infra,
		finalizer: finalizer,
	}

	request := &storeentity.StoreRequest{
		ID:          primitive.NewObjectID(),
		Name:        "Urban Finds",
		Subdomain:   "urban-finds",
		WorkspaceID: "workspace-1",
		RequestedBy: "user-1",
		Status:      storeentity.RequestStatusProvisioning,
	}
	repo.EXPECT().FindNextProvisioning(mock.Anything).Return(request, nil)
	infra.EXPECT().EnsurePlacementRoute(mock.Anything, "workspace-1", request.ID.Hex()).Return(false, nil)

	finalized, err := svc.FinalizeNextStoreRequest(context.Background())
	require.NoError(t, err)
	require.Nil(t, finalized)
}
