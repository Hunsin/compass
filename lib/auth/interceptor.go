package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type contextKey string

const (
	UserIDKey contextKey = "user_id"
)

// JWTValidator defines the interface for verifying JWT tokens
type JWTValidator interface {
	VerifyToken(ctx context.Context, tokenStr string) (*oidc.IDToken, error)
}

// keycloakValidator implements JWTValidator for Keycloak
type keycloakValidator struct {
	verifier *oidc.IDTokenVerifier
}

// NewKeycloakValidator creates a new JWT validator against Keycloak
func NewKeycloakValidator(ctx context.Context, baseURL, realm, clientID string) (JWTValidator, error) {
	baseURL = strings.TrimSuffix(baseURL, "/")
	issuerURL := fmt.Sprintf("%s/realms/%s", baseURL, realm)

	// In OIDC, we discover the provider keys via .well-known endpoints
	provider, err := oidc.NewProvider(ctx, issuerURL)
	if err != nil {
		return nil, fmt.Errorf("failed to sync provider %s: %w", issuerURL, err)
	}

	verifier := provider.Verifier(&oidc.Config{
		// Keycloak access tokens have aud "account" by default, not the client ID.
		// Skip client ID check since we are verifying access tokens, not ID tokens.
		SkipClientIDCheck: true,
	})

	return &keycloakValidator{
		verifier: verifier,
	}, nil
}

// VerifyToken verifies the raw token string
func (v *keycloakValidator) VerifyToken(ctx context.Context, tokenStr string) (*oidc.IDToken, error) {
	return v.verifier.Verify(ctx, tokenStr)
}

// HTTPMiddleware creates a middleware for validating Authorization: Bearer <token>
func HTTPMiddleware(validator JWTValidator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "authorization header missing", http.StatusUnauthorized)
				return
			}

			if !strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
				http.Error(w, "invalid authorization header format", http.StatusUnauthorized)
				return
			}

			tokenStr := authHeader[7:]
			idToken, err := validator.VerifyToken(r.Context(), tokenStr)
			if err != nil {
				http.Error(w, fmt.Sprintf("invalid token: %v", err), http.StatusUnauthorized)
				return
			}

			// Add User ID (Subject) to context
			ctx := context.WithValue(r.Context(), UserIDKey, idToken.Subject)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GRPCUnaryInterceptor creates a unary interceptor for validating JWTs in gRPC metadata
func GRPCUnaryInterceptor(validator JWTValidator, ignoreMethods ...string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		for _, m := range ignoreMethods {
			if info.FullMethod == m {
				return handler(ctx, req)
			}
		}

		ctx, err := authenticateGRPC(ctx, validator)
		if err != nil {
			return nil, err
		}

		return handler(ctx, req)
	}
}

// GRPCStreamInterceptor creates a stream interceptor for validating JWTs in gRPC metadata
func GRPCStreamInterceptor(validator JWTValidator, ignoreMethods ...string) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		for _, m := range ignoreMethods {
			if info.FullMethod == m {
				return handler(srv, ss)
			}
		}

		ctx, err := authenticateGRPC(ss.Context(), validator)
		if err != nil {
			return err
		}

		wrapped := &wrappedStream{ServerStream: ss, ctx: ctx}
		return handler(srv, wrapped)
	}
}

func authenticateGRPC(ctx context.Context, validator JWTValidator) (context.Context, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "metadata is not provided")
	}

	authValues := md["authorization"]
	if len(authValues) == 0 {
		return nil, status.Error(codes.Unauthenticated, "authorization token is not provided")
	}

	authHeader := authValues[0]
	if !strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
		return nil, status.Error(codes.Unauthenticated, "invalid authorization header format")
	}

	tokenStr := authHeader[7:]
	idToken, err := validator.VerifyToken(ctx, tokenStr)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
	}

	return context.WithValue(ctx, UserIDKey, idToken.Subject), nil
}

type wrappedStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedStream) Context() context.Context {
	return w.ctx
}

// GetUserIDFromContext retrieves the User ID stored in context by the interceptor/middleware.
func GetUserIDFromContext(ctx context.Context) (string, error) {
	val := ctx.Value(UserIDKey)
	if val == nil {
		return "", errors.New("user ID not found in context")
	}
	userID, ok := val.(string)
	if !ok {
		return "", errors.New("user ID in context is not a string")
	}
	return userID, nil
}
