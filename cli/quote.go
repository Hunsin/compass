package main

import (
	"context"
	"net"

	"github.com/jackc/pgx/v5"
	"github.com/urfave/cli/v3"
	"google.golang.org/grpc"

	"github.com/Hunsin/compass/lib/flags"
	"github.com/Hunsin/compass/postgres/gen/model"
	pb "github.com/Hunsin/compass/protocols/gen/go/quote"
	quoteservice "github.com/Hunsin/compass/services/quote"
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

			db := model.New(conn)

			lis, err := net.Listen("tcp", cmd.String("listen-addr"))
			if err != nil {
				return err
			}

			srv := grpc.NewServer()
			pb.RegisterQuoteServiceServer(srv, quoteservice.New(db))

			return srv.Serve(lis)
		},
	}
}
