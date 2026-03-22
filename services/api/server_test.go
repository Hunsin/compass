package api

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	"github.com/Hunsin/compass/lib/auth"
	pb "github.com/Hunsin/compass/protocols/gen/go/auth/v1"
)

func TestLogin(t *testing.T) {
	tests := []struct {
		name     string
		req      *pb.LoginRequest
		stub     func(*auth.MockKeycloakClient)
		wantCode codes.Code
	}{
		{
			name: "login success",
			req: &pb.LoginRequest{
				Username: proto.String("testuser"),
				Password: proto.String("testpass"),
			},
			stub: func(m *auth.MockKeycloakClient) {
				m.EXPECT().Login(mock.Anything, "testuser", "testpass").
					Return(&auth.TokenResponse{
						AccessToken:  "mock-access-token",
						RefreshToken: "mock-refresh-token",
						ExpiresIn:    300,
						TokenType:    "Bearer",
					}, nil)
			},
			wantCode: codes.OK,
		},
		{
			name: "login failure - invalid credentials",
			req: &pb.LoginRequest{
				Username: proto.String("baduser"),
				Password: proto.String("badpass"),
			},
			stub: func(m *auth.MockKeycloakClient) {
				m.EXPECT().Login(mock.Anything, "baduser", "badpass").
					Return(nil, errors.New("keycloak returned status 401: invalid credentials"))
			},
			wantCode: codes.Unauthenticated,
		},
		{
			name: "login failure - missing username",
			req: &pb.LoginRequest{
				Username: proto.String(""),
				Password: proto.String("testpass"),
			},
			wantCode: codes.InvalidArgument,
		},
		{
			name: "login failure - missing password",
			req: &pb.LoginRequest{
				Username: proto.String("testuser"),
				Password: proto.String(""),
			},
			wantCode: codes.InvalidArgument,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			kc := auth.NewMockKeycloakClient(t)
			if tc.stub != nil {
				tc.stub(kc)
			}

			srv := NewServer(kc)
			resp, err := srv.Login(context.Background(), tc.req)

			if tc.wantCode == codes.OK {
				require.NoError(t, err)
				require.Equal(t, "mock-access-token", resp.GetAccessToken())
				require.Equal(t, "mock-refresh-token", resp.GetRefreshToken())
			} else {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, tc.wantCode, st.Code())
			}
		})
	}
}

func TestMe(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		srv := NewServer(nil)

		// Simulate context with user ID (as set by auth interceptor)
		ctx := context.WithValue(context.Background(), auth.UserIDKey, "user-123")
		resp, err := srv.Me(ctx, &pb.MeRequest{})

		require.NoError(t, err)
		require.Equal(t, "user-123", resp.GetUserId())
	})

	t.Run("no user in context", func(t *testing.T) {
		srv := NewServer(nil)

		_, err := srv.Me(context.Background(), &pb.MeRequest{})
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		require.Equal(t, codes.Internal, st.Code())
	})
}
