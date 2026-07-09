package store

import (
	"context"
	"strings"
	"time"

	"go.uber.org/fx"

	infrasinputport "github.com/tuannm99/podzone/internal/onboarding/domain/infrasmanager/inputport"
	storeentity "github.com/tuannm99/podzone/internal/onboarding/domain/store/entity"
	storeinputport "github.com/tuannm99/podzone/internal/onboarding/domain/store/inputport"
	storeoutputport "github.com/tuannm99/podzone/internal/onboarding/domain/store/outputport"
	"github.com/tuannm99/podzone/pkg/collection"
	"github.com/tuannm99/podzone/pkg/toolkit"
)

var _ storeinputport.Usecase = (*StoreInteractor)(nil)

type StoreInteractor struct {
	repo        storeoutputport.StoreRepository
	authorizer  storeoutputport.AccessAuthorizer
	finalizer   storeoutputport.OperationalStoreFinalizer
	infra       infrasinputport.Usecase
	provisioner storeentity.ProvisioningConfig
}

type StoreInteractorParams struct {
	fx.In

	StoreRepo  storeoutputport.StoreRepository
	Authorizer storeoutputport.AccessAuthorizer
	Finalizer  storeoutputport.OperationalStoreFinalizer
	Infra      infrasinputport.Usecase
	Config     storeentity.ProvisioningConfig
}

func NewStoreInteractor(params StoreInteractorParams) *StoreInteractor {
	return &StoreInteractor{
		repo:        params.StoreRepo,
		authorizer:  params.Authorizer,
		finalizer:   params.Finalizer,
		infra:       params.Infra,
		provisioner: params.Config,
	}
}

func (s *StoreInteractor) CreateStoreRequest(
	ctx context.Context,
	cmd storeinputport.CreateStoreRequestCommand,
) (*storeinputport.Request, error) {
	name := strings.TrimSpace(cmd.Name)
	subdomain := strings.TrimSpace(cmd.Subdomain)
	if name == "" {
		return nil, ErrNameRequired
	}
	if subdomain == "" {
		return nil, ErrSubdomainRequired
	}
	workspaceID, err := toolkit.GetTenantID(ctx)
	if err != nil {
		return nil, ErrWorkspaceIDRequired
	}
	if workspaceID == "" {
		return nil, ErrWorkspaceIDRequired
	}
	requestedBy, err := toolkit.GetUserID(ctx)
	if err != nil {
		return nil, ErrRequestedByRequired
	}
	if requestedBy == "" {
		return nil, ErrRequestedByRequired
	}
	ownerID := strings.TrimSpace(cmd.OwnerID)
	if ownerID == "" {
		ownerID = requestedBy
	}
	if s.authorizer != nil {
		if err := s.authorizer.AuthorizeStoreRequest(ctx, workspaceID, requestedBy); err != nil {
			return nil, err
		}
	}
	if ownerID != requestedBy && s.authorizer != nil {
		if err := s.authorizer.AuthorizeStoreApproval(ctx, requestedBy); err != nil {
			return nil, err
		}
	}

	existing, err := s.repo.FindBySubdomain(ctx, subdomain)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrSubdomainTaken
	}

	now := time.Now().UTC()
	status := storeentity.RequestStatusRequested
	if s.provisioner.AutoApprove {
		status = storeentity.RequestStatusQueued
	}
	request, err := s.repo.Create(ctx, storeentity.StoreRequest{
		WorkspaceID: workspaceID,
		Name:        name,
		Subdomain:   subdomain,
		RequestedBy: requestedBy,
		OwnerID:     ownerID,
		Status:      status,
		CreatedAt:   now,
		UpdatedAt:   now,
	})
	if err != nil {
		if isDuplicateStoreError(err) {
			return nil, ErrSubdomainTaken
		}
		return nil, err
	}
	s.recordTransition(
		ctx,
		request.ID.Hex(),
		"",
		request.Status,
		map[string]string{"user": requestedBy},
		"request.created",
		"store request created",
		"",
	)

	return toInputPortRequest(request), nil
}

func (s *StoreInteractor) GetStoreRequest(ctx context.Context, id string) (*storeinputport.Request, error) {
	if err := s.authorizeRead(ctx); err != nil {
		return nil, err
	}
	request, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if request == nil {
		return nil, ErrStoreNotFound
	}
	return toInputPortRequest(request), nil
}

func (s *StoreInteractor) ListStoreRequestTransitions(
	ctx context.Context,
	id string,
	query collection.Query,
) (collection.Page[storeinputport.RequestTransition], error) {
	if err := s.authorizeRead(ctx); err != nil {
		return collection.Page[storeinputport.RequestTransition]{}, err
	}
	workspaceID, err := toolkit.GetTenantID(ctx)
	if err != nil {
		return collection.Page[storeinputport.RequestTransition]{}, ErrWorkspaceIDRequired
	}
	request, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return collection.Page[storeinputport.RequestTransition]{}, err
	}
	if request == nil || request.WorkspaceID != workspaceID {
		return collection.Page[storeinputport.RequestTransition]{}, ErrStoreNotFound
	}
	page, err := s.repo.ListTransitions(ctx, id, query)
	if err != nil {
		return collection.Page[storeinputport.RequestTransition]{}, err
	}
	items := make([]storeinputport.RequestTransition, 0, len(page.Items))
	for _, transition := range page.Items {
		items = append(items, storeinputport.RequestTransition{
			ID:        transition.ID.Hex(),
			RequestID: transition.RequestID,
			From:      storeinputport.RequestStatus(transition.From),
			To:        storeinputport.RequestStatus(transition.To),
			Actor:     transition.Actor,
			Step:      transition.Step,
			Reason:    transition.Reason,
			ErrorCode: transition.ErrorCode,
			CreatedAt: transition.CreatedAt,
		})
	}
	return collection.Page[storeinputport.RequestTransition]{
		Items:       items,
		Total:       page.Total,
		Page:        page.Page,
		PageSize:    page.PageSize,
		TotalPages:  page.TotalPages,
		HasNext:     page.HasNext,
		HasPrevious: page.HasPrevious,
	}, nil
}

func (s *StoreInteractor) ApproveStoreRequest(ctx context.Context, id string) error {
	if err := s.authorizeApproval(ctx); err != nil {
		return err
	}
	return s.UpdateStoreRequestStatus(ctx, id, storeinputport.RequestStatusQueued)
}

func (s *StoreInteractor) RejectStoreRequest(ctx context.Context, id string) error {
	if err := s.authorizeApproval(ctx); err != nil {
		return err
	}
	return s.UpdateStoreRequestStatus(ctx, id, storeinputport.RequestStatusRejected)
}

func (s *StoreInteractor) RetryStoreRequest(ctx context.Context, id string) error {
	workspaceID, err := toolkit.GetTenantID(ctx)
	if err != nil {
		return ErrWorkspaceIDRequired
	}
	requestedBy, err := toolkit.GetUserID(ctx)
	if err != nil {
		return ErrRequestedByRequired
	}
	request, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if request == nil || request.WorkspaceID != workspaceID {
		return ErrStoreNotFound
	}
	if s.authorizer != nil {
		if err := s.authorizer.AuthorizeStoreRequest(ctx, workspaceID, requestedBy); err != nil {
			return err
		}
	}
	if !isValidStatusTransition(
		storeinputport.RequestStatus(request.Status),
		storeinputport.RequestStatusQueued,
	) {
		return ErrInvalidStatus
	}
	if err := s.repo.UpdateStatus(ctx, id, storeentity.RequestStatusQueued); err != nil {
		return err
	}
	s.recordTransition(
		ctx,
		id,
		request.Status,
		storeentity.RequestStatusQueued,
		map[string]string{"user": requestedBy},
		"request.retried",
		"retry requested",
		"",
	)
	return nil
}

func (s *StoreInteractor) ListStoreRequests(
	ctx context.Context,
	workspaceID string,
	query collection.Query,
) (collection.Page[*storeinputport.Request], error) {
	requestedBy, err := toolkit.GetUserID(ctx)
	if err != nil {
		return collection.Page[*storeinputport.Request]{}, ErrRequestedByRequired
	}
	if s.authorizer != nil {
		if err := s.authorizer.AuthorizeStoreRead(ctx, workspaceID, requestedBy); err != nil {
			return collection.Page[*storeinputport.Request]{}, err
		}
	}
	page, err := s.repo.ListPage(ctx, workspaceID, query.Normalize())
	if err != nil {
		return collection.Page[*storeinputport.Request]{}, err
	}

	out := make([]*storeinputport.Request, 0, len(page.Items))
	for _, req := range page.Items {
		copyReq := req
		out = append(out, toInputPortRequest(&copyReq))
	}
	return collection.NewPage(out, page.Total, query), nil
}

func (s *StoreInteractor) UpdateStoreRequestStatus(
	ctx context.Context,
	id string,
	status storeinputport.RequestStatus,
) error {
	current, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if current == nil {
		return ErrStoreNotFound
	}

	if !isValidStatusTransition(storeinputport.RequestStatus(current.Status), status) {
		return ErrInvalidStatus
	}

	next := storeentity.RequestStatus(status)
	if err := s.repo.UpdateStatus(ctx, id, next); err != nil {
		return err
	}
	s.recordTransition(
		ctx,
		id,
		current.Status,
		next,
		map[string]string{"system": "onboarding"},
		statusTransitionStep(next),
		"status updated",
		"",
	)
	return nil
}

func isValidStatusTransition(current, next storeinputport.RequestStatus) bool {
	switch current {
	case "":
		return next == storeinputport.RequestStatusRequested || next == storeinputport.RequestStatusPendingApproval
	case storeinputport.RequestStatusRequested:
		return next == storeinputport.RequestStatusPendingApproval || next == storeinputport.RequestStatusQueued
	case storeinputport.RequestStatusPendingApproval:
		return next == storeinputport.RequestStatusQueued || next == storeinputport.RequestStatusRejected
	case storeinputport.RequestStatusQueued:
		return next == storeinputport.RequestStatusPlanning ||
			next == storeinputport.RequestStatusProvisioning ||
			next == storeinputport.RequestStatusFailed
	case storeinputport.RequestStatusPlanning:
		return next == storeinputport.RequestStatusPlanned ||
			next == storeinputport.RequestStatusPendingPlatformSetup ||
			next == storeinputport.RequestStatusFailedRetryable ||
			next == storeinputport.RequestStatusFailedNonRetryable ||
			next == storeinputport.RequestStatusFailed
	case storeinputport.RequestStatusPlanned:
		return next == storeinputport.RequestStatusProvisioning ||
			next == storeinputport.RequestStatusPendingApproval ||
			next == storeinputport.RequestStatusPendingPlatformSetup
	case storeinputport.RequestStatusProvisioning:
		return next == storeinputport.RequestStatusReady ||
			next == storeinputport.RequestStatusFailed ||
			next == storeinputport.RequestStatusFailedRetryable ||
			next == storeinputport.RequestStatusFailedNonRetryable ||
			next == storeinputport.RequestStatusPendingPlatformSetup
	case storeinputport.RequestStatusReady:
		return next == storeinputport.RequestStatusSuspended || next == storeinputport.RequestStatusArchived
	case storeinputport.RequestStatusFailed,
		storeinputport.RequestStatusFailedRetryable,
		storeinputport.RequestStatusPendingPlatformSetup:
		return next == storeinputport.RequestStatusQueued || next == storeinputport.RequestStatusPendingApproval
	case storeinputport.RequestStatusFailedNonRetryable:
		return next == storeinputport.RequestStatusPendingApproval
	case storeinputport.RequestStatusRejected:
		return next == storeinputport.RequestStatusQueued
	case storeinputport.RequestStatusSuspended:
		return next == storeinputport.RequestStatusReady
	case storeinputport.RequestStatusArchived:
		return false
	default:
		return false
	}
}

func (s *StoreInteractor) ProcessNextStoreRequest(
	ctx context.Context,
	workerID string,
) (*storeinputport.Request, error) {
	if !s.provisioner.Enabled {
		return nil, ErrProvisionerDisabled
	}
	if s.infra == nil {
		return nil, ErrProvisionerMissing
	}

	workerID = normalizeWorkerID(workerID)
	request, err := s.repo.ClaimNextQueued(ctx, workerID, s.leaseTTL())
	if err != nil {
		return nil, err
	}
	if request == nil {
		return nil, nil
	}
	s.recordTransition(
		ctx,
		request.ID.Hex(),
		request.Status,
		storeentity.RequestStatusPlanning,
		map[string]string{"service": "onboarding", "worker": "store-provisioner"},
		"planning.started",
		"worker claimed queued request",
		"",
	)

	storeID := request.ID.Hex()
	_, err = s.infra.ProvisionStorePlacement(
		ctx,
		infrasinputport.ProvisionStorePlacementRequest{
			RequestID:   request.ID.Hex(),
			TenantID:    request.WorkspaceID,
			StoreID:     storeID,
			Subdomain:   request.Subdomain,
			RequestedBy: request.RequestedBy,
		},
		map[string]string{
			"service": "onboarding",
			"worker":  "store-provisioner",
		},
	)
	if err != nil {
		status, errorCode := classifyProvisioningFailure(err)
		_ = s.repo.MarkBlocked(ctx, request.ID.Hex(), status, err.Error())
		s.recordTransition(
			ctx,
			request.ID.Hex(),
			storeentity.RequestStatusPlanning,
			status,
			map[string]string{"service": "onboarding", "worker": "store-provisioner"},
			"provisioning.failed",
			err.Error(),
			errorCode,
		)
		return nil, err
	}

	if err := s.repo.UpdateStatus(ctx, request.ID.Hex(), storeentity.RequestStatusPlanned); err != nil {
		return nil, err
	}
	s.recordTransition(
		ctx,
		request.ID.Hex(),
		storeentity.RequestStatusPlanning,
		storeentity.RequestStatusPlanned,
		map[string]string{"service": "onboarding", "worker": "store-provisioner"},
		"planning.completed",
		"placement plan persisted and provisioning completed",
		"",
	)
	if err := s.repo.UpdateStatus(ctx, request.ID.Hex(), storeentity.RequestStatusProvisioning); err != nil {
		return nil, err
	}
	s.recordTransition(
		ctx,
		request.ID.Hex(),
		storeentity.RequestStatusPlanned,
		storeentity.RequestStatusProvisioning,
		map[string]string{"service": "onboarding", "worker": "store-provisioner"},
		"provisioning.started",
		"waiting for route readiness and store finalization",
		"",
	)
	request.Status = storeentity.RequestStatusProvisioning
	return toInputPortRequest(request), nil
}

func (s *StoreInteractor) FinalizeNextStoreRequest(
	ctx context.Context,
	workerID string,
) (*storeinputport.Request, error) {
	if s.finalizer == nil {
		return nil, ErrProvisionerMissing
	}
	workerID = normalizeWorkerID(workerID)
	request, err := s.repo.ClaimNextProvisioning(ctx, workerID, s.leaseTTL())
	if err != nil || request == nil {
		return nil, err
	}
	if s.infra != nil {
		ready, err := s.infra.EnsurePlacementRoute(ctx, request.WorkspaceID)
		if err != nil {
			return nil, err
		}
		if !ready {
			if err := s.repo.ReleaseLease(ctx, request.ID.Hex(), workerID); err != nil {
				return nil, err
			}
			return nil, ErrRouteNotReady
		}
		s.recordTransition(
			ctx,
			request.ID.Hex(),
			storeentity.RequestStatusProvisioning,
			storeentity.RequestStatusProvisioning,
			map[string]string{"service": "onboarding", "worker": "store-provisioner"},
			"route.ready",
			"tenant placement route is ready",
			"",
		)
	}
	if err := s.finalizer.FinalizeStore(ctx, *request); err != nil {
		return nil, err
	}
	s.recordTransition(
		ctx,
		request.ID.Hex(),
		storeentity.RequestStatusProvisioning,
		storeentity.RequestStatusProvisioning,
		map[string]string{"service": "onboarding", "worker": "store-provisioner"},
		"store.finalized",
		"operational store finalized",
		"",
	)
	storeID := request.ID.Hex()
	if err := s.repo.MarkReady(ctx, request.ID.Hex(), storeID); err != nil {
		return nil, err
	}
	s.recordTransition(
		ctx,
		request.ID.Hex(),
		storeentity.RequestStatusProvisioning,
		storeentity.RequestStatusReady,
		map[string]string{"service": "onboarding", "worker": "store-provisioner"},
		"request.ready",
		"store provisioning completed",
		"",
	)
	request.Status = storeentity.RequestStatusReady
	request.StoreID = &request.ID
	return toInputPortRequest(request), nil
}

func (s *StoreInteractor) authorizeApproval(ctx context.Context) error {
	requestedBy, err := toolkit.GetUserID(ctx)
	if err != nil {
		return ErrRequestedByRequired
	}
	if s.authorizer == nil {
		return nil
	}
	return s.authorizer.AuthorizeStoreApproval(ctx, requestedBy)
}

func (s *StoreInteractor) authorizeRead(ctx context.Context) error {
	workspaceID, err := toolkit.GetTenantID(ctx)
	if err != nil {
		return ErrWorkspaceIDRequired
	}
	requestedBy, err := toolkit.GetUserID(ctx)
	if err != nil {
		return ErrRequestedByRequired
	}
	if s.authorizer == nil {
		return nil
	}
	return s.authorizer.AuthorizeStoreRead(ctx, workspaceID, requestedBy)
}

func isDuplicateStoreError(err error) bool {
	return strings.Contains(strings.ToLower(err.Error()), "duplicate")
}

func classifyProvisioningFailure(err error) (storeentity.RequestStatus, string) {
	if err == nil {
		return storeentity.RequestStatusFailedRetryable, ""
	}
	message := strings.ToLower(err.Error())
	switch {
	case strings.Contains(message, "not implemented"),
		strings.Contains(message, "resource inventory"),
		strings.Contains(message, "capacity unavailable"),
		strings.Contains(message, "admin_dsn is required"),
		strings.Contains(message, "cluster not found"),
		strings.Contains(message, "namespace not found"),
		strings.Contains(message, "runtime pool not found"):
		return storeentity.RequestStatusPendingPlatformSetup, "platform_setup_required"
	case strings.Contains(message, "invalid"),
		strings.Contains(message, "unsupported placement runtime"),
		strings.Contains(message, "unsupported postgres placement mode"):
		return storeentity.RequestStatusFailedNonRetryable, "non_retryable"
	default:
		return storeentity.RequestStatusFailedRetryable, "retryable"
	}
}

func (s *StoreInteractor) recordTransition(
	ctx context.Context,
	requestID string,
	from storeentity.RequestStatus,
	to storeentity.RequestStatus,
	actor map[string]string,
	step string,
	reason string,
	errorCode string,
) {
	if s.repo == nil {
		return
	}
	_ = s.repo.RecordTransition(ctx, storeentity.StoreRequestTransition{
		RequestID: requestID,
		From:      from,
		To:        to,
		Actor:     actor,
		Step:      step,
		Reason:    reason,
		ErrorCode: errorCode,
		CreatedAt: time.Now().UTC(),
	})
}

func normalizeWorkerID(workerID string) string {
	if workerID == "" {
		return "store-provisioner"
	}
	return workerID
}

func (s *StoreInteractor) leaseTTL() time.Duration {
	if s.provisioner.LeaseTTL <= 0 {
		return 2 * time.Minute
	}
	return s.provisioner.LeaseTTL
}

func statusTransitionStep(status storeentity.RequestStatus) string {
	switch status {
	case storeentity.RequestStatusQueued:
		return "approval.queued"
	case storeentity.RequestStatusRejected:
		return "approval.rejected"
	default:
		return "status.updated"
	}
}
