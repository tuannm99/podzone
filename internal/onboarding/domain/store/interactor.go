package store

import (
	"context"
	"strings"
	"time"

	"go.uber.org/fx"

	storeentity "github.com/tuannm99/podzone/internal/onboarding/domain/store/entity"
	storeinputport "github.com/tuannm99/podzone/internal/onboarding/domain/store/inputport"
	storeoutputport "github.com/tuannm99/podzone/internal/onboarding/domain/store/outputport"
)

var _ storeinputport.Usecase = (*StoreInteractor)(nil)

type StoreInteractor struct {
	repo storeoutputport.StoreRepository
}

type StoreInteractorParams struct {
	fx.In

	StoreRepo storeoutputport.StoreRepository
}

func NewStoreInteractor(params StoreInteractorParams) *StoreInteractor {
	return &StoreInteractor{repo: params.StoreRepo}
}

func (s *StoreInteractor) CreateStoreRequest(
	ctx context.Context,
	name, subdomain, workspaceID, requestedBy string,
) (*storeinputport.Request, error) {
	if name == "" {
		return nil, ErrNameRequired
	}
	if subdomain == "" {
		return nil, ErrSubdomainRequired
	}
	if workspaceID == "" {
		return nil, ErrWorkspaceIDRequired
	}
	if requestedBy == "" {
		return nil, ErrRequestedByRequired
	}

	existing, err := s.repo.FindBySubdomain(ctx, subdomain)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrSubdomainTaken
	}

	now := time.Now().UTC()
	request, err := s.repo.Create(ctx, storeentity.StoreRequest{
		WorkspaceID: workspaceID,
		Name:        name,
		Subdomain:   subdomain,
		RequestedBy: requestedBy,
		Status:      storeentity.RequestStatusRequested,
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
	return s.UpdateStoreRequestStatus(ctx, id, storeinputport.RequestStatusQueued)
}

func (s *StoreInteractor) RejectStoreRequest(ctx context.Context, id string) error {
	return s.UpdateStoreRequestStatus(ctx, id, storeinputport.RequestStatusRejected)
}

func (s *StoreInteractor) ListStoreRequests(
	ctx context.Context,
	workspaceID string,
) ([]*storeinputport.Request, error) {
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

func (s *StoreInteractor) CreateStore(
	ctx context.Context,
	name, subdomain, ownerID string,
) (*storeinputport.Request, error) {
	return s.CreateStoreRequest(ctx, name, subdomain, ownerID, ownerID)
}

func (s *StoreInteractor) GetStore(ctx context.Context, id string) (*storeinputport.Request, error) {
	return s.GetStoreRequest(ctx, id)
}

func (s *StoreInteractor) GetStoresByOwner(ctx context.Context, ownerID string) ([]*storeinputport.Request, error) {
	return s.ListStoreRequests(ctx, ownerID)
}

func (s *StoreInteractor) UpdateStoreStatus(ctx context.Context, id string, status storeinputport.StoreStatus) error {
	return s.UpdateStoreRequestStatus(ctx, id, status)
}

func isDuplicateStoreError(err error) bool {
	return strings.Contains(strings.ToLower(err.Error()), "duplicate")
}
