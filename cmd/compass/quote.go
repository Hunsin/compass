package main

import (
	"context"
	"net"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/urfave/cli/v3"
	"google.golang.org/grpc"

	"github.com/Hunsin/compass/lib/flags"
	quoteLib "github.com/Hunsin/compass/lib/quote"
	pb "github.com/Hunsin/compass/protocols/gen/go/quote"
	quoteSvc "github.com/Hunsin/compass/services/quote"
)

func quoteCommand() *cli.Command {
	return &cli.Command{
		Name:  "quote",
		Usage: "Start the Quote gRPC service",
		Flags: []cli.Flag{&flags.PostgresURL, &flags.ListenAddr},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			conn, err := pgx.Connect(ctx, cmd.String("postgres-url"))
			if err != nil {
				return err
			}
			defer conn.Close(ctx)

			childCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()
			if err := conn.Ping(childCtx); err != nil {
				return err
			}

			lis, err := net.Listen("tcp", cmd.String("listen-addr"))
			if err != nil {
				return err
			}

			srv := grpc.NewServer()
			model := quoteLib.Connect(conn)
			pb.RegisterQuoteServiceServer(srv, quoteSvc.New(model))

			return srv.Serve(lis)
		},
	}
}
