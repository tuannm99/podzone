package mapper

import (
	"github.com/tuannm99/podzone/internal/auth/domain/inputport"
	pbauthv1 "github.com/tuannm99/podzone/pkg/api/proto/auth/v1"
)

func ToPBGoogleCallbackResponse(
	out *inputport.GoogleCallbackResult,
) *pbauthv1.GoogleCallbackResponse {
	return &pbauthv1.GoogleCallbackResponse{
		JwtToken:    out.JwtToken,
		RedirectUrl: out.RedirectUrl,
		UserInfo:    ToPBGoogleUserInfo(&out.UserInfo),
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
