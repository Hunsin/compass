package main

import (
	"context"
	"net/http"

	"github.com/urfave/cli/v3"

	"github.com/Hunsin/compass/lib/auth"
	"github.com/Hunsin/compass/lib/flags"
	"github.com/Hunsin/compass/services/api"
)

func apiCommand() *cli.Command {
	return &cli.Command{
		Name:  "api",
		Usage: "Start the Frontend JSON API server",
		Flags: []cli.Flag{
			&flags.ListenAddr,
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
			listenAddr := cmd.String(flags.ListenAddr.Name)

			// 1. Initialize Keycloak HTTP Client (for Login)
			kcClient := auth.NewKeycloakClient(kcURL, kcRealm, kcClientID, kcClientSecret)

			// 2. Initialize JWT Validator (for protected routes)
			validator, err := auth.NewKeycloakValidator(ctx, kcURL, kcRealm, kcClientID)
			if err != nil {
				return err
			}

			// 3. Initialize and Start API Server
			server := api.NewServer(kcClient, validator)

			return http.ListenAndServe(listenAddr, server)
		},
	}
}
