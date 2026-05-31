package worker

import (
	"context"
	"errors"
	"time"

	onboardingconfig "github.com/tuannm99/podzone/internal/onboarding/config"
	storeinputport "github.com/tuannm99/podzone/internal/onboarding/domain/store/inputport"
	"github.com/tuannm99/podzone/pkg/pdlog"
)

type StoreProvisioningWorker struct {
	log      pdlog.Logger
	store    storeinputport.Usecase
	enabled  bool
	interval time.Duration
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

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

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
	request, err := w.store.ProcessNextStoreRequest(ctx)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return
		}
		w.log.Error("onboarding store provisioning tick failed", "error", err)
		return
	}
	if request != nil {
		w.log.Info(
			"onboarding store request provisioned",
			"request_id", request.ID,
			"workspace_id", request.WorkspaceID,
			"store_id", request.StoreID,
		)
	}
}
