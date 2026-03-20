package api

import (
	"context"
	"log"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	"github.com/Hunsin/compass/lib/auth"
	pb "github.com/Hunsin/compass/protocols/gen/go/auth/v1"
)

// Server implements the gRPC AuthService.
type Server struct {
	pb.UnimplementedAuthServiceServer
	kcClient auth.KeycloakClient
}

// NewServer creates a new AuthService gRPC server.
func NewServer(kcClient auth.KeycloakClient) *Server {
	return &Server{
		kcClient: kcClient,
	}
}

// Login authenticates a user via Keycloak and returns OAuth2 tokens.
func (s *Server) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	if req.GetUsername() == "" || req.GetPassword() == "" {
		return nil, status.Error(codes.InvalidArgument, "username and password are required")
	}

	tokenResp, err := s.kcClient.Login(ctx, req.GetUsername(), req.GetPassword())
	if err != nil {
		log.Printf("Login failed for user %s: %v", req.GetUsername(), err)
		return nil, status.Error(codes.Unauthenticated, "invalid credentials or upstream error")
	}

	return &pb.LoginResponse{
		AccessToken:      proto.String(tokenResp.AccessToken),
		ExpiresIn:        proto.Int32(int32(tokenResp.ExpiresIn)),
		RefreshExpiresIn: proto.Int32(int32(tokenResp.RefreshExpiresIn)),
		RefreshToken:     proto.String(tokenResp.RefreshToken),
		TokenType:        proto.String(tokenResp.TokenType),
		IdToken:          proto.String(tokenResp.IdToken),
		SessionState:     proto.String(tokenResp.SessionState),
		Scope:            proto.String(tokenResp.Scope),
	}, nil
}

// Me returns the authenticated user's information from the JWT token.
func (s *Server) Me(ctx context.Context, _ *pb.MeRequest) (*pb.MeResponse, error) {
	userID, err := auth.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get user info")
	}

	return &pb.MeResponse{
		UserId: proto.String(userID),
	}, nil
}
