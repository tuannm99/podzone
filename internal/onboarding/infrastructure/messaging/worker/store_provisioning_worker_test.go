package worker

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"

	storeinputport "github.com/tuannm99/podzone/internal/onboarding/domain/store/inputport"
	storemocks "github.com/tuannm99/podzone/internal/onboarding/domain/store/inputport/mocks"
	logmocks "github.com/tuannm99/podzone/pkg/pdlog/mocks"
)

func TestStoreProvisioningWorkerTickProcessesBatchesWithWorkerLease(t *testing.T) {
	log := logmocks.NewMockLogger(t)
	store := storemocks.NewMockUsecase(t)
	worker := &StoreProvisioningWorker{
		log:      log,
		store:    store,
		batch:    3,
		workerID: "worker-test",
	}

	store.EXPECT().
		ReclaimNextStalePlanningRequest(mock.Anything).
		Return(nil, nil).
		Once()
	store.EXPECT().
		ProcessNextStoreRequest(mock.Anything, "worker-test").
		Return(&storeinputport.Request{ID: "request-1", WorkspaceID: "workspace-1"}, nil).
		Once()
	store.EXPECT().
		ProcessNextStoreRequest(mock.Anything, "worker-test").
		Return(nil, nil).
		Once()
	store.EXPECT().
		FinalizeNextStoreRequest(mock.Anything, "worker-test").
		Return(nil, nil).
		Once()
	log.EXPECT().Info("onboarding store placement requested", mock.Anything).Return().Once()

	worker.tick(context.Background())
}

// P7's residual gap (docs/07-problems/onboarding-provisioning-2026-07-07.md):
// the reclaim sweep runs ahead of claiming new work, and stops the batch
// loop once nothing more is stale.
func TestStoreProvisioningWorkerTickReclaimsStalePlanningRequests(t *testing.T) {
	log := logmocks.NewMockLogger(t)
	store := storemocks.NewMockUsecase(t)
	worker := &StoreProvisioningWorker{
		log:      log,
		store:    store,
		batch:    3,
		workerID: "worker-test",
	}

	store.EXPECT().
		ReclaimNextStalePlanningRequest(mock.Anything).
		Return(&storeinputport.Request{ID: "request-1", WorkspaceID: "workspace-1"}, nil).
		Once()
	store.EXPECT().
		ReclaimNextStalePlanningRequest(mock.Anything).
		Return(nil, nil).
		Once()
	store.EXPECT().
		ProcessNextStoreRequest(mock.Anything, "worker-test").
		Return(nil, nil).
		Once()
	store.EXPECT().
		FinalizeNextStoreRequest(mock.Anything, "worker-test").
		Return(nil, nil).
		Once()
	log.EXPECT().Info("onboarding stale planning request requeued", mock.Anything).Return().Once()

	worker.tick(context.Background())
}
