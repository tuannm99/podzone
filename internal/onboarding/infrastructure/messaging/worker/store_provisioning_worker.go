package worker

import (
	"context"
	"errors"
	"os"
	"strconv"
	"time"

	onboardingconfig "github.com/tuannm99/podzone/internal/onboarding/config"
	storedomain "github.com/tuannm99/podzone/internal/onboarding/domain/store"
	storeinputport "github.com/tuannm99/podzone/internal/onboarding/domain/store/inputport"
	"github.com/tuannm99/podzone/pkg/pdlog"
)

type StoreProvisioningWorker struct {
	log      pdlog.Logger
	store    storeinputport.Usecase
	enabled  bool
	interval time.Duration
	batch    int
	workerID string
}

func NewStoreProvisioningWorker(
	log pdlog.Logger,
	store storeinputport.Usecase,
	cfg onboardingconfig.StoreProvisioningConfig,
) *StoreProvisioningWorker {
	return &StoreProvisioningWorker{
		log:      log,
		store:    store,
		enabled:  cfg.Enabled,
		interval: cfg.Interval,
		batch:    cfg.BatchSize,
		workerID: newStoreProvisioningWorkerID(),
	}
}

func (w *StoreProvisioningWorker) Run(ctx context.Context) {
	if !w.enabled {
		w.log.Info("Onboarding store provisioning worker disabled")
		return
	}
	if w.interval <= 0 {
		w.interval = 5 * time.Second
	}
	if w.batch <= 0 {
		w.batch = 5
	}

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	w.tick(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.tick(ctx)
		}
	}
}

func (w *StoreProvisioningWorker) tick(ctx context.Context) {
	for i := 0; i < w.batch; i++ {
		request, err := w.store.ProcessNextStoreRequest(ctx, w.workerID)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return
			}
			w.log.Error("onboarding store provisioning tick failed", "error", err)
			return
		}
		if request == nil {
			break
		}
		w.log.Info(
			"onboarding store placement requested",
			"request_id", request.ID,
			"workspace_id", request.WorkspaceID,
		)
	}

	for i := 0; i < w.batch; i++ {
		request, err := w.store.FinalizeNextStoreRequest(ctx, w.workerID)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return
			}
			if errors.Is(err, storedomain.ErrRouteNotReady) {
				continue
			}
			w.log.Error("onboarding store finalization tick failed", "error", err)
			return
		}
		if request == nil {
			break
		}
		w.log.Info(
			"onboarding store request ready",
			"request_id", request.ID,
			"workspace_id", request.WorkspaceID,
			"store_id", request.StoreID,
		)
	}
}

func newStoreProvisioningWorkerID() string {
	hostname, err := os.Hostname()
	if err != nil || hostname == "" {
		hostname = "unknown-host"
	}
	return "store-provisioner:" + hostname + ":" + strconv.Itoa(os.Getpid())
}
