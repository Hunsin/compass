package main

import (
	"context"
	"net"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/urfave/cli/v3"
	"google.golang.org/grpc"

	"github.com/Hunsin/compass/lib/flags"
	quoteLib "github.com/Hunsin/compass/lib/quote"
	pb "github.com/Hunsin/compass/protocols/gen/go/quote/v1"
	quoteSvc "github.com/Hunsin/compass/services/quote"
)

func quoteCommand() *cli.Command {
	return &cli.Command{
		Name:  "quote",
		Usage: "Start the Quote gRPC service",
		Flags: []cli.Flag{&flags.PostgresURL, &flags.RedisURL, &flags.ListenAddr},
		Action: func(ctx context.Context, cmd *cli.Command) error {

			// initalize Postgres and Redis clients
			pool, err := pgxpool.New(ctx, cmd.String(flags.PostgresURL.Name))
			if err != nil {
				return err
			}
			defer pool.Close()

			rdbOpts, err := redis.ParseURL(cmd.String(flags.RedisURL.Name))
			if err != nil {
				return err
			}
			rdb := redis.NewClient(rdbOpts)
			defer rdb.Close()

			// check connections
			childCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()
			if err := pool.Ping(childCtx); err != nil {
				return err
			}
			_, err = rdb.Ping(childCtx).Result()
			if err != nil {
				return err
			}

			lis, err := net.Listen("tcp", cmd.String(flags.ListenAddr.Name))
			if err != nil {
				return err
			}

			srv := grpc.NewServer()
			model := quoteLib.Connect(pool, rdb)
			pb.RegisterQuoteServiceServer(srv, quoteSvc.New(model))

			return srv.Serve(lis)
		},
	}
}
