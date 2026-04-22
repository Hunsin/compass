package main

import (
	"context"
	"net"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/rs/zerolog"
	"github.com/urfave/cli/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"

	"github.com/Hunsin/compass/lib/auth"
	"github.com/Hunsin/compass/lib/flags"
	pb "github.com/Hunsin/compass/protocols/gen/go/auth/v1"
	"github.com/Hunsin/compass/services/api"
)

func apiCommand() *cli.Command {
	return &cli.Command{
		Name:  "api",
		Usage: "Start the Auth gRPC + JSON API server (gRPC + grpc-gateway)",
		Flags: []cli.Flag{
			&flags.GRPCAddr,
			&flags.HTTPAddr,
			&flags.KeycloakURL,
			&flags.KeycloakRealm,
			&flags.KeycloakClientID,
			&flags.KeycloakClientSecret,
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			kcURL := cmd.String(flags.KeycloakURL.Name)
			kcRealm := cmd.String(flags.KeycloakRealm.Name)
			kcClientID := cmd.String(flags.KeycloakClientID.Name)
			kcClientSecret := cmd.String(flags.KeycloakClientSecret.Name)

			// 1. Initialize Keycloak HTTP Client (for Login)
			kcClient := auth.NewKeycloakClient(kcURL, kcRealm, kcClientID, kcClientSecret)

			// 2. Initialize JWT Validator (for protected routes)
			validator, err := auth.NewKeycloakValidator(ctx, kcURL, kcRealm, kcClientID)
			if err != nil {
				return err
			}

			// 3. Create gRPC server with auth interceptor (Login is excluded from auth)
			ignoreMethods := []string{pb.AuthService_Login_FullMethodName, healthpb.Health_Check_FullMethodName}
			grpcSrv := grpc.NewServer(
				grpc.ChainUnaryInterceptor(
					auth.GRPCUnaryInterceptor(validator, ignoreMethods...),
				),
			)
			svc := api.NewServer(kcClient)
			pb.RegisterAuthServiceServer(grpcSrv, svc)

			hs := health.NewServer()
			hs.SetServingStatus(pb.AuthService_ServiceDesc.ServiceName, healthpb.HealthCheckResponse_SERVING)
			hs.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
			healthpb.RegisterHealthServer(grpcSrv, hs)

			// 4. Start gRPC listener
			grpcAddr := cmd.String(flags.GRPCAddr.Name)
			lis, err := net.Listen("tcp", grpcAddr)
			if err != nil {
				return err
			}
			log := zerolog.Ctx(ctx)
			go func() {
				log.Info().Str("addr", grpcAddr).Msg("gRPC server listening")
				if err := grpcSrv.Serve(lis); err != nil {
					log.Fatal().Err(err).Msg("gRPC server error")
				}
			}()

			// 5. Create grpc-gateway mux and register in-process handler
			//
			// NOTE: In-process mode (RegisterAuthServiceHandlerServer) bypasses gRPC
			// interceptors. JWT authentication for HTTP requests is handled by
			// auth.HTTPMiddleware wrapping the gateway mux below.
			gwMux := runtime.NewServeMux()
			if err := pb.RegisterAuthServiceHandlerServer(ctx, gwMux, svc); err != nil {
				return err
			}

			// 6. Start HTTP listener (JSON gateway) with auth middleware
			// Login endpoint is public; all other routes require a valid JWT.
			httpAddr := cmd.String(flags.HTTPAddr.Name)
			log.Info().Str("addr", httpAddr).Msg("HTTP gateway listening")
			handler := authMiddlewareWithExclusions(
				auth.HTTPMiddleware(validator),
				gwMux,
				"/api/login",
			)
			return http.ListenAndServe(httpAddr, handler)
		},
	}
}

// authMiddlewareWithExclusions wraps a handler with auth middleware, but
// skips authentication for the specified paths.
func authMiddlewareWithExclusions(
	middleware func(http.Handler) http.Handler,
	handler http.Handler,
	excludedPaths ...string,
) http.Handler {
	protected := middleware(handler)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, path := range excludedPaths {
			if r.URL.Path == path {
				handler.ServeHTTP(w, r)
				return
			}
		}
		protected.ServeHTTP(w, r)
	})
}
