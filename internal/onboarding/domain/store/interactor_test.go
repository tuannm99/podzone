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
	"github.com/tuannm99/podzone/pkg/collection"
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
				request.OwnerID == "1" &&
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

func TestCreateStoreRequest_AuthorizesOwnerOverride(t *testing.T) {
	repo := storemocks.NewMockStoreRepository(t)
	allowTransitionLog(repo)
	authorizer := storemocks.NewMockAccessAuthorizer(t)
	svc := &StoreInteractor{repo: repo, authorizer: authorizer}
	ctx := authenticatedContext("workspace-1", "platform-admin")

	authorizer.EXPECT().
		AuthorizeStoreRequest(mock.Anything, "workspace-1", "platform-admin").
		Return(nil)
	authorizer.EXPECT().
		AuthorizeStoreApproval(mock.Anything, "platform-admin").
		Return(nil)
	repo.EXPECT().FindBySubdomain(mock.Anything, "new").Return(nil, nil)
	repo.EXPECT().
		Create(mock.Anything, mock.MatchedBy(func(request storeentity.StoreRequest) bool {
			return request.RequestedBy == "platform-admin" && request.OwnerID == "tenant-root"
		})).
		RunAndReturn(func(_ context.Context, request storeentity.StoreRequest) (*storeentity.StoreRequest, error) {
			request.ID = primitive.NewObjectID()
			return &request, nil
		})

	request, err := svc.CreateStoreRequest(
		ctx,
		storeinputport.CreateStoreRequestCommand{
			Name:      "New",
			Subdomain: "new",
			OwnerID:   "tenant-root",
		},
	)

	require.NoError(t, err)
	require.Equal(t, "tenant-root", request.OwnerID)
}

func TestCreateStoreRequest_PermitsOwnerOverrideWithoutAuthorizer(t *testing.T) {
	repo := storemocks.NewMockStoreRepository(t)
	allowTransitionLog(repo)
	svc := &StoreInteractor{repo: repo}

	repo.EXPECT().FindBySubdomain(mock.Anything, "new").Return(nil, nil)
	repo.EXPECT().Create(mock.Anything, mock.Anything).Return(&storeentity.StoreRequest{}, nil)

	_, err := svc.CreateStoreRequest(
		authenticatedContext("workspace-1", "user-1"),
		storeinputport.CreateStoreRequestCommand{
			Name:      "New",
			Subdomain: "new",
			OwnerID:   "another-user",
		},
	)
	require.NoError(t, err)
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

func TestListStoreRequestsReturnsRepositoryPage(t *testing.T) {
	t.Parallel()

	svc, repo := setupStoreInteractor(t)
	query := collection.Query{
		Page:          2,
		PageSize:      5,
		Search:        "urban",
		SortDirection: collection.SortDescending,
	}
	request := storeentity.StoreRequest{
		ID:          primitive.NewObjectID(),
		WorkspaceID: "workspace-1",
		Name:        "Urban Finds",
		Status:      storeentity.RequestStatusQueued,
	}
	repo.EXPECT().
		ListPage(mock.Anything, "workspace-1", query).
		Return(collection.NewPage([]storeentity.StoreRequest{request}, 6, query), nil)

	page, err := svc.ListStoreRequests(
		authenticatedContext("workspace-1", "7"),
		"workspace-1",
		query,
	)

	require.NoError(t, err)
	require.Len(t, page.Items, 1)
	require.Equal(t, request.ID.Hex(), page.Items[0].ID)
	require.Equal(t, int64(6), page.Total)
	require.True(t, page.HasPrevious)
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
	repo.EXPECT().
		ClaimNextQueued(mock.Anything, "worker-1", 2*time.Minute).
		Return(request, nil)
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
	processed, err := svc.ProcessNextStoreRequest(context.Background(), "worker-1")
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

	repo.EXPECT().
		ClaimNextQueued(mock.Anything, "worker-1", 2*time.Minute).
		Return(request, nil)
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

	processed, err := svc.ProcessNextStoreRequest(context.Background(), "worker-1")

	require.ErrorIs(t, err, providerErr)
	require.Nil(t, processed)
}

func TestFinalizeNextStoreRequest_BootstrapsOperationalStoreBeforeReady(t *testing.T) {
	repo := storemocks.NewMockStoreRepository(t)
	transitions := make([]storeentity.StoreRequestTransition, 0, 3)
	repo.EXPECT().
		RecordTransition(mock.Anything, mock.Anything).
		RunAndReturn(func(_ context.Context, transition storeentity.StoreRequestTransition) error {
			transitions = append(transitions, transition)
			return nil
		}).
		Times(3)
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
	repo.EXPECT().
		ClaimNextProvisioning(mock.Anything, "worker-1", 2*time.Minute).
		Return(request, nil)
	infra.EXPECT().EnsurePlacementRoute(mock.Anything, "workspace-1").Return(true, nil)
	finalizer.EXPECT().FinalizeStore(mock.Anything, *request).Return(nil)
	repo.EXPECT().MarkReady(mock.Anything, id.Hex(), id.Hex()).Return(nil)

	finalized, err := svc.FinalizeNextStoreRequest(context.Background(), "worker-1")
	require.NoError(t, err)
	require.NotNil(t, finalized)
	require.Equal(t, storeinputport.RequestStatusReady, finalized.Status)
	require.Equal(t, id.Hex(), finalized.StoreID)
	require.Equal(t, []string{"route.ready", "store.finalized", "request.ready"}, []string{
		transitions[0].Step,
		transitions[1].Step,
		transitions[2].Step,
	})
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
	repo.EXPECT().
		ClaimNextProvisioning(mock.Anything, "worker-1", 2*time.Minute).
		Return(request, nil)
	infra.EXPECT().EnsurePlacementRoute(mock.Anything, "workspace-1").Return(false, nil)
	repo.EXPECT().ReleaseLease(mock.Anything, request.ID.Hex(), "worker-1").Return(nil)

	finalized, err := svc.FinalizeNextStoreRequest(context.Background(), "worker-1")
	require.ErrorIs(t, err, ErrRouteNotReady)
	require.Nil(t, finalized)
}

func makeReadinessRequest(workspaceID string) *storeentity.StoreRequest {
	return &storeentity.StoreRequest{
		ID:          primitive.NewObjectID(),
		Name:        "Urban Finds",
		Subdomain:   "urban-finds",
		WorkspaceID: workspaceID,
		RequestedBy: "user-1",
	}
}

func TestGetStoreReadiness_Ready(t *testing.T) {
	repo := storemocks.NewMockStoreRepository(t)
	infra := infrasmocks.NewMockUsecase(t)
	svc := &StoreInteractor{repo: repo, infra: infra}

	req := makeReadinessRequest("workspace-1")
	req.Status = storeentity.RequestStatusReady
	repo.EXPECT().FindByID(mock.Anything, req.ID.Hex()).Return(req, nil)
	infra.EXPECT().GetTenantPlacementStatus(mock.Anything, "workspace-1").
		Return(&infrasinputport.PlacementStatus{AllocationReady: true, RouteReady: true}, nil)

	ctx := authenticatedContext("workspace-1", "user-1")
	resp, err := svc.GetStoreReadiness(ctx, req.ID.Hex())
	require.NoError(t, err)
	require.Equal(t, storeinputport.UIStateReady, resp.UIState)
	require.True(t, resp.Readiness.StoreReady)
	require.True(t, resp.Readiness.PlacementAllocationReady)
	require.True(t, resp.Readiness.RouteReady)
}

func TestGetStoreReadiness_BlockedWhenPlacementNotReady(t *testing.T) {
	repo := storemocks.NewMockStoreRepository(t)
	infra := infrasmocks.NewMockUsecase(t)
	svc := &StoreInteractor{repo: repo, infra: infra}

	req := makeReadinessRequest("workspace-1")
	req.Status = storeentity.RequestStatusReady
	repo.EXPECT().FindByID(mock.Anything, req.ID.Hex()).Return(req, nil)
	infra.EXPECT().GetTenantPlacementStatus(mock.Anything, "workspace-1").
		Return(&infrasinputport.PlacementStatus{AllocationReady: false, RouteReady: false}, nil)

	ctx := authenticatedContext("workspace-1", "user-1")
	resp, err := svc.GetStoreReadiness(ctx, req.ID.Hex())
	require.NoError(t, err)
	require.Equal(t, storeinputport.UIStateBlocked, resp.UIState)
	require.True(t, resp.Readiness.StoreReady)
	require.False(t, resp.Readiness.PlacementAllocationReady)
}

func TestGetStoreReadiness_Provisioning(t *testing.T) {
	repo := storemocks.NewMockStoreRepository(t)
	svc := &StoreInteractor{repo: repo}

	req := makeReadinessRequest("workspace-1")
	req.Status = storeentity.RequestStatusProvisioning
	repo.EXPECT().FindByID(mock.Anything, req.ID.Hex()).Return(req, nil)

	ctx := authenticatedContext("workspace-1", "user-1")
	resp, err := svc.GetStoreReadiness(ctx, req.ID.Hex())
	require.NoError(t, err)
	require.Equal(t, storeinputport.UIStateProvisioning, resp.UIState)
	require.False(t, resp.Readiness.StoreReady)
}

func TestGetStoreReadiness_Failed(t *testing.T) {
	repo := storemocks.NewMockStoreRepository(t)
	svc := &StoreInteractor{repo: repo}

	req := makeReadinessRequest("workspace-1")
	req.Status = storeentity.RequestStatusFailed
	req.LastError = "provisioning timed out"
	repo.EXPECT().FindByID(mock.Anything, req.ID.Hex()).Return(req, nil)

	ctx := authenticatedContext("workspace-1", "user-1")
	resp, err := svc.GetStoreReadiness(ctx, req.ID.Hex())
	require.NoError(t, err)
	require.Equal(t, storeinputport.UIStateFailed, resp.UIState)
	require.Equal(t, "provisioning timed out", resp.FailureReason)
}

func TestGetStoreReadiness_NotFound(t *testing.T) {
	repo := storemocks.NewMockStoreRepository(t)
	svc := &StoreInteractor{repo: repo}

	repo.EXPECT().FindByID(mock.Anything, "missing-id").Return(nil, nil)

	ctx := authenticatedContext("workspace-1", "user-1")
	_, err := svc.GetStoreReadiness(ctx, "missing-id")
	require.ErrorIs(t, err, ErrStoreNotFound)
}

func TestGetStoreReadiness_WorkspaceMismatchReturns404(t *testing.T) {
	repo := storemocks.NewMockStoreRepository(t)
	svc := &StoreInteractor{repo: repo}

	req := makeReadinessRequest("workspace-other")
	req.Status = storeentity.RequestStatusReady
	repo.EXPECT().FindByID(mock.Anything, req.ID.Hex()).Return(req, nil)

	ctx := authenticatedContext("workspace-1", "user-1")
	_, err := svc.GetStoreReadiness(ctx, req.ID.Hex())
	require.ErrorIs(t, err, ErrStoreNotFound)
}

func TestGetStoreReadiness_NoAuth(t *testing.T) {
	svc := &StoreInteractor{}

	_, err := svc.GetStoreReadiness(context.Background(), "some-id")
	require.ErrorIs(t, err, ErrWorkspaceIDRequired)
}
