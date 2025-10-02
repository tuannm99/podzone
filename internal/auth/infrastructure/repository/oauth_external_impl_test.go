package repository

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func TestGoogleOauthImpl_FetchUserInfo_WithLocalServer(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"email":  "jdoe@example.com",
			"name":   "John Doe",
			"locale": "vi",
		})
	}))
	defer ts.Close()

	_ = NewGoogleOauthImpl() // sonar pass
	g := NewGoogleOauthImplWithOptions(&oauth2.Config{}, http.DefaultClient, ts.URL)
	ui, err := g.FetchUserInfo("fake-token")
	require.NoError(t, err)
	require.NotNil(t, ui)
	require.Equal(t, "jdoe@example.com", ui.Email)
}
