package store

import (
	storeentity "github.com/tuannm99/podzone/internal/onboarding/domain/store/entity"
	storeinputport "github.com/tuannm99/podzone/internal/onboarding/domain/store/inputport"
)

func toInputPortRequest(req *storeentity.StoreRequest) *storeinputport.Request {
	if req == nil {
		return nil
	}
	out := &storeinputport.Request{
		ID:          req.ID.Hex(),
		WorkspaceID: req.WorkspaceID,
		Name:        req.Name,
		Subdomain:   req.Subdomain,
		RequestedBy: req.RequestedBy,
		OwnerID:     storeOwnerID(req),
		Status:      storeinputport.RequestStatus(req.Status),
		LastError:   req.LastError,
		CreatedAt:   req.CreatedAt,
		UpdatedAt:   req.UpdatedAt,
	}
	if req.StoreID != nil {
		out.StoreID = req.StoreID.Hex()
	}
	if req.ApprovedAt != nil {
		approvedAt := *req.ApprovedAt
		out.ApprovedAt = &approvedAt
	}
	if req.CompletedAt != nil {
		completedAt := *req.CompletedAt
		out.CompletedAt = &completedAt
	}
	return out
}

func storeOwnerID(req *storeentity.StoreRequest) string {
	if req.OwnerID != "" {
		return req.OwnerID
	}
	return req.RequestedBy
}
