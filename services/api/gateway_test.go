package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/Hunsin/compass/lib/auth"
	pb "github.com/Hunsin/compass/protocols/gen/go/auth/v1"
)

// newTestGateway creates an httptest.Server that mirrors the real API server wiring:
// grpc-gateway mux + HTTPMiddleware with /api/login excluded from auth.
func newTestGateway(t *testing.T, kc auth.KeycloakClient, validator auth.JWTValidator) *httptest.Server {
	t.Helper()

	svc := NewServer(kc)

	gwMux := runtime.NewServeMux()
	require.NoError(t, pb.RegisterAuthServiceHandlerServer(t.Context(), gwMux, svc))

	// Mirror the middleware setup from cmd/compass/api.go
	protected := auth.HTTPMiddleware(validator)(gwMux)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/login" {
			gwMux.ServeHTTP(w, r)
			return
		}
		protected.ServeHTTP(w, r)
	})

	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return srv
}

func TestGatewayLogin(t *testing.T) {
	kc := auth.NewMockKeycloakClient(t)
	kc.EXPECT().Login(mock.Anything, "testuser", "testpass").
		Return(&auth.TokenResponse{
			AccessToken:  "mock-access-token",
			RefreshToken: "mock-refresh-token",
			ExpiresIn:    300,
			TokenType:    "Bearer",
		}, nil)

	srv := newTestGateway(t, kc, nil)

	// Login should succeed without any Authorization header (public endpoint).
	resp, err := http.Post(
		srv.URL+"/api/login",
		"application/json",
		strings.NewReader(`{"username":"testuser","password":"testpass"}`),
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var body map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	require.Equal(t, "mock-access-token", body["accessToken"])
}

func TestGatewayMe(t *testing.T) {
	validator := auth.NewMockJWTValidator(t)

	t.Run("success", func(t *testing.T) {
		validator.EXPECT().VerifyToken(mock.Anything, "valid-token").
			Return(&oidc.IDToken{Subject: "user-123"}, nil).Once()

		srv := newTestGateway(t, nil, validator)

		req, err := http.NewRequest(http.MethodGet, srv.URL+"/api/me", nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer valid-token")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)

		var body map[string]any
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
		require.Equal(t, "user-123", body["userId"])
	})

	t.Run("missing token", func(t *testing.T) {
		srv := newTestGateway(t, nil, validator)

		resp, err := http.Get(srv.URL + "/api/me")
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}
