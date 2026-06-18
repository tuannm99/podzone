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
	if s.authorizer != nil {
		if err := s.authorizer.AuthorizeStoreRequest(ctx, workspaceID, requestedBy); err != nil {
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
	return s.repo.UpdateStatus(ctx, id, storeentity.RequestStatusQueued)
}

func (s *StoreInteractor) ListStoreRequests(
	ctx context.Context,
	workspaceID string,
) ([]*storeinputport.Request, error) {
	requestedBy, err := toolkit.GetUserID(ctx)
	if err != nil {
		return nil, ErrRequestedByRequired
	}
	if s.authorizer != nil {
		if err := s.authorizer.AuthorizeStoreRead(ctx, workspaceID, requestedBy); err != nil {
			return nil, err
		}
	}
	requests, err := s.repo.List(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	out := make([]*storeinputport.Request, 0, len(requests))
	for _, req := range requests {
		copyReq := req
		out = append(out, toInputPortRequest(&copyReq))
	}
	return out, nil
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

	return s.repo.UpdateStatus(ctx, id, storeentity.RequestStatus(status))
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
		return next == storeinputport.RequestStatusProvisioning || next == storeinputport.RequestStatusFailed
	case storeinputport.RequestStatusProvisioning:
		return next == storeinputport.RequestStatusReady || next == storeinputport.RequestStatusFailed
	case storeinputport.RequestStatusReady:
		return next == storeinputport.RequestStatusSuspended || next == storeinputport.RequestStatusArchived
	case storeinputport.RequestStatusFailed:
		return next == storeinputport.RequestStatusQueued || next == storeinputport.RequestStatusPendingApproval
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

func (s *StoreInteractor) ProcessNextStoreRequest(ctx context.Context) (*storeinputport.Request, error) {
	if !s.provisioner.Enabled {
		return nil, ErrProvisionerDisabled
	}
	if s.infra == nil {
		return nil, ErrProvisionerMissing
	}

	request, err := s.repo.ClaimNextQueued(ctx)
	if err != nil {
		return nil, err
	}
	if request == nil {
		return nil, nil
	}

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
		_ = s.repo.MarkFailed(ctx, request.ID.Hex(), err.Error())
		return nil, err
	}

	return toInputPortRequest(request), nil
}

func (s *StoreInteractor) FinalizeNextStoreRequest(ctx context.Context) (*storeinputport.Request, error) {
	if s.finalizer == nil {
		return nil, ErrProvisionerMissing
	}
	request, err := s.repo.FindNextProvisioning(ctx)
	if err != nil || request == nil {
		return nil, err
	}
	if s.infra != nil {
		ready, err := s.infra.EnsurePlacementRoute(ctx, request.WorkspaceID, request.ID.Hex())
		if err != nil {
			return nil, err
		}
		if !ready {
			return nil, nil
		}
	}
	if err := s.finalizer.FinalizeStore(ctx, *request); err != nil {
		return nil, err
	}
	storeID := request.ID.Hex()
	if err := s.repo.MarkReady(ctx, request.ID.Hex(), storeID); err != nil {
		return nil, err
	}
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
