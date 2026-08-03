package pdglobalmiddleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"github.com/tuannm99/podzone/pkg/pdlog"
)

func TestNewReadyRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)

	newEngine := func(checker ReadinessChecker) *gin.Engine {
		r := gin.New()
		NewReadyRoute(pdlog.NopLogger{}, checker)(r)
		return r
	}

	t.Run("nil checker is always ready", func(t *testing.T) {
		r := newEngine(nil)
		req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("checker returning nil is ready", func(t *testing.T) {
		r := newEngine(func(ctx context.Context) error { return nil })
		req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("checker returning an error is not ready", func(t *testing.T) {
		r := newEngine(func(ctx context.Context) error { return errors.New("mongo unreachable") })
		req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusServiceUnavailable, rec.Code)
		require.Contains(t, rec.Body.String(), "mongo unreachable")
	})
}
