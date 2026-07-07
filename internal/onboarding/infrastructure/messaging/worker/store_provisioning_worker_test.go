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
