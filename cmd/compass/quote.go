package main

import (
	"context"
	"net"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/urfave/cli/v3"
	"google.golang.org/grpc"

	"github.com/Hunsin/compass/lib/auth"
	"github.com/Hunsin/compass/lib/flags"
	quoteLib "github.com/Hunsin/compass/lib/quote"
	pb "github.com/Hunsin/compass/protocols/gen/go/quote/v1"
	quoteSvc "github.com/Hunsin/compass/services/quote"
)

func quoteCommand() *cli.Command {
	return &cli.Command{
		Name:  "quote",
		Usage: "Start the Quote gRPC service",
		Flags: []cli.Flag{
			&flags.PostgresURL,
			&flags.RedisURL,
			&flags.GRPCAddr,
			&flags.KeycloakURL,
			&flags.KeycloakRealm,
			&flags.KeycloakClientID,
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {

			// initialize the Postgres client
			pool, err := pgxpool.New(ctx, cmd.String(flags.PostgresURL.Name))
			if err != nil {
				return err
			}
			defer pool.Close()

			// check connection
			pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()
			if err := pool.Ping(pingCtx); err != nil {
				return err
			}

			// Redis is optional; create connection if the flag is provided
			var rdb *redis.Client
			if u := cmd.String(flags.RedisURL.Name); u != "" {
				opts, err := redis.ParseURL(u)
				if err != nil {
					return err
				}

				rdb = redis.NewClient(opts)
				defer rdb.Close()

				pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
				defer cancel()
				if _, err := rdb.Ping(pingCtx).Result(); err != nil {
					return err
				}
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

			srv := grpc.NewServer(
				grpc.UnaryInterceptor(auth.GRPCUnaryInterceptor(validator)),
				grpc.StreamInterceptor(auth.GRPCStreamInterceptor(validator)),
			)
			model := quoteLib.Connect(pool, rdb)
			pb.RegisterQuoteServiceServer(srv, quoteSvc.New(model))

			return srv.Serve(lis)
		},
	}
}
