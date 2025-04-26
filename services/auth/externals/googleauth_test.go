package externals_test

import (
	"errors"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tuannm99/podzone/services/auth/externals"
	"go.uber.org/zap"
)

func TestFetchUserInfo(t *testing.T) {
	httpmock.Activate(t)
	defer httpmock.DeactivateAndReset()

	accessTokenMock := "access_token"
	requestUrl := "https://www.googleapis.com/oauth2/v3/userinfo?access_token=" + accessTokenMock

	googleExternal := externals.NewGoogleOauthExternal(zap.NewExample())

	t.Run("network error", func(t *testing.T) {
		httpmock.RegisterResponder("GET", requestUrl,
			httpmock.NewErrorResponder(errors.New("network failure")),
		)

		_, err := googleExternal.FetchUserInfo(accessTokenMock)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "network failure")
	})

	t.Run("bad json response", func(t *testing.T) {
		httpmock.RegisterResponder("GET", requestUrl,
			httpmock.NewStringResponder(200, "bad_json"),
		)

		_, err := googleExternal.FetchUserInfo(accessTokenMock)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid character")
	})

	t.Run("successful response", func(t *testing.T) {
		mockBody := `{
			"email": "test@example.com",
			"name": "Test User",
			"sub": "12345"
		}`

		httpmock.RegisterResponder("GET", requestUrl,
			httpmock.NewStringResponder(200, mockBody),
		)

		info, err := googleExternal.FetchUserInfo(accessTokenMock)
		require.NoError(t, err)
		assert.NotNil(t, info)
		assert.Equal(t, "test@example.com", info.Email)
		assert.Equal(t, "Test User", info.Name)
		assert.Equal(t, "12345", info.Sub)
	})
}
