package mapper

import (
	"time"

	"github.com/tuannm99/podzone/internal/auth/domain/entity"
	"github.com/tuannm99/podzone/internal/auth/domain/inputport"
	pbauthv1 "github.com/tuannm99/podzone/pkg/api/proto/auth/v1"
	"github.com/tuannm99/podzone/pkg/toolkit"
)

func ToPBGoogleCallbackResponse(
	out *inputport.GoogleCallbackResult,
) *pbauthv1.GoogleCallbackResponse {
	return &pbauthv1.GoogleCallbackResponse{
		ExchangeCode: out.ExchangeCode,
		RedirectUrl:  out.RedirectUrl,
		UserInfo:     ToPBGoogleUserInfo(&out.UserInfo),
	}
}

func ToPBGoogleUserInfo(in *inputport.GoogleUserInfo) *pbauthv1.GoogleUserInfo {
	return &pbauthv1.GoogleUserInfo{
		Id:            in.Id,
		Email:         in.Email,
		Name:          in.Name,
		GivenName:     in.GivenName,
		FamilyName:    in.FamilyName,
		Picture:       in.Picture,
		EmailVerified: in.EmailVerified,
	}
}

func ToPBUserInfo(in *entity.User) *pbauthv1.UserInfo {
	if in == nil {
		return nil
	}
	return &pbauthv1.UserInfo{
		Id:          int32(in.Id),
		Email:       in.Email,
		Username:    in.Username,
		FullName:    in.FullName,
		MiddleName:  in.MiddleName,
		FirstName:   in.FirstName,
		LastName:    in.LastName,
		Address:     in.Address,
		InitialFrom: in.InitialFrom,
		Age:         int32(in.Age),
		Dob:         in.Dob.Format(time.RFC3339),
	}
}

func ToPBLoginResponse(out *inputport.AuthResult) (*pbauthv1.LoginResponse, error) {
	if out == nil {
		return nil, nil
	}
	return toolkit.MapStruct[inputport.AuthResult, pbauthv1.LoginResponse](*out)
}

func ToPBRegisterResponse(out *inputport.AuthResult) (*pbauthv1.RegisterResponse, error) {
	if out == nil {
		return nil, nil
	}
	return toolkit.MapStruct[inputport.AuthResult, pbauthv1.RegisterResponse](*out)
}

func ToPBSwitchActiveTenantResponse(
	out *inputport.AuthResult,
) (*pbauthv1.SwitchActiveTenantResponse, error) {
	if out == nil {
		return nil, nil
	}
	return toolkit.MapStruct[inputport.AuthResult, pbauthv1.SwitchActiveTenantResponse](*out)
}

func ToPBRefreshTokenResponse(out *inputport.AuthResult) (*pbauthv1.RefreshTokenResponse, error) {
	if out == nil {
		return nil, nil
	}
	return toolkit.MapStruct[inputport.AuthResult, pbauthv1.RefreshTokenResponse](*out)
}

func ToRegisterCmd(in *pbauthv1.RegisterRequest) (*inputport.RegisterCmd, error) {
	if in == nil {
		return nil, nil
	}
	return toolkit.MapStruct[*pbauthv1.RegisterRequest, inputport.RegisterCmd](in)
}

func ToPBSession(s *entity.Session) *pbauthv1.Session {
	if s == nil {
		return nil
	}
	resp := &pbauthv1.Session{
		Id:                          s.ID,
		UserId:                      uint64(s.UserID),
		ActiveTenantId:              s.ActiveTenantID,
		Status:                      s.Status,
		CreatedAt:                   s.CreatedAt.Format(time.RFC3339),
		UpdatedAt:                   s.UpdatedAt.Format(time.RFC3339),
		ExpiresAt:                   s.ExpiresAt.Format(time.RFC3339),
		SessionPolicy:               ToPBSessionPolicyStatements(s.SessionPolicy),
		AssumedRoleId:               s.AssumedRoleID,
		AssumedRoleScope:            s.AssumedRoleScope,
		AssumedRoleName:             s.AssumedRoleName,
		AssumedRoleTenantId:         s.AssumedRoleTenantID,
		AssumedRoleServicePrincipal: s.AssumedRoleServicePrincipal,
		AssumedRoleSessionName:      s.AssumedRoleSessionName,
		AssumedRoleSourceIdentity:   s.AssumedRoleSourceIdentity,
		SessionTags:                 cloneStringMap(s.SessionTags),
	}
	if s.AssumedRoleExpiresAt != nil {
		resp.AssumedRoleExpiresAt = s.AssumedRoleExpiresAt.Format(time.RFC3339)
	}
	if s.RevokedAt != nil {
		resp.RevokedAt = s.RevokedAt.Format(time.RFC3339)
	}
	return resp
}

func ToPBAuditLog(a *entity.AuditLog) *pbauthv1.AuditLog {
	if a == nil {
		return nil
	}
	return &pbauthv1.AuditLog{
		Id:           a.ID,
		ActorUserId:  uint64(a.ActorUserID),
		Action:       a.Action,
		ResourceType: a.ResourceType,
		ResourceId:   a.ResourceID,
		TenantId:     a.TenantID,
		Status:       a.Status,
		PayloadJson:  a.PayloadJSON,
		CreatedAt:    a.CreatedAt.Format(time.RFC3339),
	}
}

func ToPBSessionPolicyStatements(items []entity.SessionPolicyStatement) []*pbauthv1.PolicyStatement {
	out := make([]*pbauthv1.PolicyStatement, 0, len(items))
	for _, item := range items {
		out = append(out, &pbauthv1.PolicyStatement{
			Effect:          item.Effect,
			ActionPattern:   item.ActionPattern,
			ResourcePattern: item.ResourcePattern,
			Conditions:      toPBSessionPolicyConditions(item.Conditions),
		})
	}
	return out
}

func FromPBSessionPolicyStatements(items []*pbauthv1.PolicyStatement) []entity.SessionPolicyStatement {
	out := make([]entity.SessionPolicyStatement, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		out = append(out, entity.SessionPolicyStatement{
			Effect:          item.Effect,
			ActionPattern:   item.ActionPattern,
			ResourcePattern: item.ResourcePattern,
			Conditions:      fromPBSessionPolicyConditions(item.Conditions),
		})
	}
	return out
}

func toPBSessionPolicyConditions(items []entity.SessionPolicyCondition) []*pbauthv1.PolicyCondition {
	out := make([]*pbauthv1.PolicyCondition, 0, len(items))
	for _, item := range items {
		out = append(out, &pbauthv1.PolicyCondition{
			Operator: item.Operator,
			Key:      item.Key,
			Value:    item.Value,
		})
	}
	return out
}

func fromPBSessionPolicyConditions(items []*pbauthv1.PolicyCondition) []entity.SessionPolicyCondition {
	out := make([]entity.SessionPolicyCondition, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		out = append(out, entity.SessionPolicyCondition{
			Operator: item.Operator,
			Key:      item.Key,
			Value:    item.Value,
		})
	}
	return out
}

func cloneStringMap(src map[string]string) map[string]string {
	if len(src) == 0 {
		return nil
	}
	out := make(map[string]string, len(src))
	for k, v := range src {
		out[k] = v
	}
	return out
}
