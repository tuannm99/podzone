package httphandler

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"

	onboardingconfig "github.com/tuannm99/podzone/internal/onboarding/config"
	"github.com/tuannm99/podzone/pkg/pdauthn"
	"github.com/tuannm99/podzone/pkg/toolkit"
)

func TestAuthentication_UsesTenantHeaderWhenTokenHasNoActiveTenant(t *testing.T) {
	recorder := performAuthenticatedRequest(t, "", "tenant-1")

	require.Equal(t, http.StatusNoContent, recorder.Code)
	require.Equal(t, "tenant-1", recorder.Header().Get("X-Test-Tenant-ID"))
}

func TestAuthentication_RejectsTenantHeaderThatDiffersFromToken(t *testing.T) {
	recorder := performAuthenticatedRequest(t, "tenant-1", "tenant-2")

	require.Equal(t, http.StatusForbidden, recorder.Code)
	require.JSONEq(t, `{"error":"tenant header does not match active session"}`, recorder.Body.String())
}

func performAuthenticatedRequest(
	t *testing.T,
	activeTenantID string,
	requestedTenantID string,
) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	const secret = "test-secret"
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, pdauthn.Claims{
		UserID:         7,
		ActiveTenantID: activeTenantID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	})
	signedToken, err := token.SignedString([]byte(secret))
	require.NoError(t, err)

	auth := NewAuthentication(onboardingconfig.AuthConfig{
		Authn: pdauthn.Config{JWTSecret: secret},
	})
	router := gin.New()
	router.Use(auth.RequireUser)
	router.GET("/requests", func(ctx *gin.Context) {
		tenantID, tenantErr := toolkit.GetTenantID(ctx.Request.Context())
		require.NoError(t, tenantErr)
		ctx.Header("X-Test-Tenant-ID", tenantID)
		ctx.Status(http.StatusNoContent)
	})

	request := httptest.NewRequest(http.MethodGet, "/requests", nil)
	request.Header.Set("Authorization", "Bearer "+signedToken)
	request.Header.Set("X-Tenant-ID", requestedTenantID)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	return recorder
}
