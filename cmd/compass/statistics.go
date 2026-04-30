package main

import (
	"context"
	"net"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"github.com/urfave/cli/v3"
	"google.golang.org/grpc"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"

	"github.com/Hunsin/compass/lib/auth"
	"github.com/Hunsin/compass/lib/flags"
	"github.com/Hunsin/compass/lib/middleware"
	statsLib "github.com/Hunsin/compass/lib/statistics"
	pb "github.com/Hunsin/compass/protocols/gen/go/statistics/v1"
	statsSvc "github.com/Hunsin/compass/services/statistics"
)

func statisticsCommand() *cli.Command {
	return &cli.Command{
		Name:  "statistics",
		Usage: "Start the Statistics gRPC service",
		Flags: []cli.Flag{
			&flags.PostgresURL,
			&flags.GRPCAddr,
			&flags.KeycloakURL,
			&flags.KeycloakRealm,
			&flags.KeycloakClientID,
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			pool, err := pgxpool.New(ctx, cmd.String(flags.PostgresURL.Name))
			if err != nil {
				return err
			}
			defer pool.Close()

			pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()
			if err := pool.Ping(pingCtx); err != nil {
				return err
			}

			lis, err := net.Listen("tcp", cmd.String(flags.GRPCAddr.Name))
			if err != nil {
				return err
			}

			kcURL := cmd.String(flags.KeycloakURL.Name)
			kcRealm := cmd.String(flags.KeycloakRealm.Name)
			kcClientID := cmd.String(flags.KeycloakClientID.Name)

			validator, err := auth.NewKeycloakValidator(ctx, kcURL, kcRealm, kcClientID)
			if err != nil {
				return err
			}

			log := zerolog.Ctx(ctx)
			srv := grpc.NewServer(
				grpc.ChainUnaryInterceptor(
					auth.GRPCUnaryInterceptor(validator, healthpb.Health_Check_FullMethodName),
					middleware.UnaryInterceptor(log),
				),
				grpc.ChainStreamInterceptor(
					auth.GRPCStreamInterceptor(validator, healthpb.Health_Watch_FullMethodName),
					middleware.StreamInterceptor(log),
				),
			)
			handler := statsSvc.New(statsLib.Connect(pool))
			pb.RegisterStatisticsServiceServer(srv, handler)
			healthpb.RegisterHealthServer(srv, handler)

			log.Info().Str("addr", lis.Addr().String()).Msg("starting statistics service")
			return srv.Serve(lis)
		},
	}
}
