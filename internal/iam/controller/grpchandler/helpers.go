package grpchandler

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
	iammapper "github.com/tuannm99/podzone/internal/iam/controller/mapper"
	iamdomain "github.com/tuannm99/podzone/internal/iam/domain/entity"
	"github.com/tuannm99/podzone/pkg/collection"
	"github.com/tuannm99/podzone/pkg/pdauthn"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func toUint(v uint64) (uint, error) {
	if v == 0 {
		return 0, status.Error(codes.InvalidArgument, "user_id is required")
	}
	if v > math.MaxUint {
		return 0, status.Error(codes.InvalidArgument, "user_id is out of range")
	}
	return uint(v), nil
}

func mergeStringMaps(items ...map[string]string) map[string]string {
	merged := make(map[string]string)
	for _, item := range items {
		for k, v := range item {
			merged[k] = v
		}
	}
	if len(merged) == 0 {
		return nil
	}
	return merged
}

func toPrincipalTagAttributes(tags map[string]string) map[string]string {
	if len(tags) == 0 {
		return nil
	}
	out := make(map[string]string, len(tags)*2)
	for k, v := range tags {
		out["principal_tag:"+k] = v
		out["request_tag:"+k] = v
	}
	return out
}

func inviteAcceptBaseURL(appRedirectURL string) string {
	base := strings.TrimSpace(appRedirectURL)
	if base == "" {
		return ""
	}
	base = strings.TrimRight(base, "/")
	base = strings.TrimSuffix(base, "/auth/google/callback")
	return base
}

func (s *iamHandlerBase) recordAudit(
	ctx context.Context,
	actorUserID uint,
	action string,
	resourceType string,
	resourceID string,
	tenantID string,
	payload map[string]any,
) {
	if s.auditRep == nil {
		return
	}
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		payloadJSON = []byte("{}")
	}
	_ = s.auditRep.Create(ctx, iamdomain.AuditLog{
		ID:           uuid.NewString(),
		ActorUserID:  actorUserID,
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		TenantID:     tenantID,
		Status:       "success",
		PayloadJSON:  string(payloadJSON),
		CreatedAt:    time.Now().UTC(),
	})
}

func iamStatusError(err error) error {
	switch {
	case errors.Is(err, iamdomain.ErrInvalidTenantName),
		errors.Is(err, collection.ErrInvalidQuery),
		errors.Is(err, iamdomain.ErrInvalidTenantSlug),
		errors.Is(err, iamdomain.ErrInvalidOrganizationName),
		errors.Is(err, iamdomain.ErrInvalidOrganizationSlug),
		errors.Is(err, iamdomain.ErrInvalidUserID),
		errors.Is(err, iamdomain.ErrInvalidRoleName),
		errors.Is(err, iamdomain.ErrInvalidPolicyName),
		errors.Is(err, iamdomain.ErrInvalidAssumeRole),
		errors.Is(err, iamdomain.ErrInvalidServicePrincipal),
		errors.Is(err, iamdomain.ErrInvalidPolicyStatement):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, iamdomain.ErrTenantNotFound),
		errors.Is(err, iamdomain.ErrOrganizationNotFound),
		errors.Is(err, iamdomain.ErrMembershipNotFound),
		errors.Is(err, iamdomain.ErrRoleNotFound),
		errors.Is(err, iamdomain.ErrPolicyNotFound),
		errors.Is(err, iamdomain.ErrPolicyVersionNotFound),
		errors.Is(err, iamdomain.ErrGroupNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, iamdomain.ErrTenantSlugTaken):
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, iamdomain.ErrPermissionDenied),
		errors.Is(err, iamdomain.ErrAssumeRoleDenied):
		return status.Error(codes.PermissionDenied, err.Error())
	case errors.Is(err, iamdomain.ErrInactiveMembership):
		return status.Error(codes.FailedPrecondition, err.Error())
	case errors.Is(err, iamdomain.ErrImmutablePolicy),
		errors.Is(err, iamdomain.ErrImmutableGroup),
		errors.Is(err, iamdomain.ErrPolicyInUse),
		errors.Is(err, iamdomain.ErrDefaultPolicyVersion):
		return status.Error(codes.FailedPrecondition, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}

func (s *iamHandlerBase) authorizedContext(ctx context.Context) (context.Context, uint, error) {
	claims, err := s.claimsFromContext(ctx)
	if err != nil {
		return ctx, 0, err
	}
	ctx = iamdomain.WithSessionPolicyStatements(ctx, iammapper.ToIAMSessionPolicyStatements(claims.SessionPolicy))
	ctx = iamdomain.WithSessionTags(ctx, claims.SessionTags)
	if claims.AssumedRoleID != 0 && claims.AssumedRoleName != "" {
		ctx = iamdomain.WithAssumedRole(ctx, iamdomain.AssumedRole{
			RoleID:           claims.AssumedRoleID,
			RoleScope:        claims.AssumedRoleScope,
			RoleName:         claims.AssumedRoleName,
			TenantID:         claims.AssumedRoleTenantID,
			ServicePrincipal: claims.AssumedRoleServicePrincipal,
			SessionName:      claims.AssumedRoleSessionName,
			SourceIdentity:   claims.AssumedRoleSourceIdentity,
			SessionTags:      claims.SessionTags,
		})
	}
	if claims.UserID == 0 {
		return ctx, 0, errors.New("access token missing user_id")
	}
	return ctx, claims.UserID, nil
}

func (s *iamHandlerBase) claimsFromContext(ctx context.Context) (*pdauthn.Claims, error) {
	return s.verifier.ClaimsFromContext(ctx)
}

func (s *iamHandlerBase) claimsFromAccessToken(accessToken string) (*pdauthn.Claims, error) {
	return s.verifier.ClaimsFromAccessToken(accessToken)
}
