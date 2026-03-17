package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/Hunsin/compass/lib/auth"
)

func TestHandleLogin(t *testing.T) {
	tests := []struct {
		name       string
		body       loginRequest
		stub       func(*auth.MockKeycloakClient)
		wantStatus int
	}{
		{
			name: "login success",
			body: loginRequest{Username: "testuser", Password: "testpass"},
			stub: func(m *auth.MockKeycloakClient) {
				m.EXPECT().Login(mock.Anything, "testuser", "testpass").
					Return(&auth.TokenResponse{
						AccessToken:  "mock-access-token",
						RefreshToken: "mock-refresh-token",
						ExpiresIn:    300,
						TokenType:    "Bearer",
					}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "login failure - invalid credentials",
			body: loginRequest{Username: "baduser", Password: "badpass"},
			stub: func(m *auth.MockKeycloakClient) {
				m.EXPECT().Login(mock.Anything, "baduser", "badpass").
					Return(nil, errors.New("keycloak returned status 401: invalid credentials"))
			},
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			kc := auth.NewMockKeycloakClient(t)
			if tc.stub != nil {
				tc.stub(kc)
			}

			srv := NewServer(kc, nil)

			body, err := json.Marshal(tc.body)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			srv.ServeHTTP(rec, req)

			require.Equal(t, tc.wantStatus, rec.Code)

			if tc.wantStatus == http.StatusOK {
				var resp auth.TokenResponse
				err := json.NewDecoder(rec.Body).Decode(&resp)
				require.NoError(t, err)
				require.Equal(t, "mock-access-token", resp.AccessToken)
				require.Equal(t, "mock-refresh-token", resp.RefreshToken)
			}
		})
	}
}
