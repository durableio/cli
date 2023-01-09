/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"strings"

	"github.com/durableio/cli/pkg/durable"
	"github.com/durableio/cli/pkg/http/server"
	"github.com/durableio/cli/pkg/logging"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// devCmd represents the dev command
var devCmd = &cobra.Command{
	Use:   "dev",
	Short: "Run durable locally",
	Run: func(cmd *cobra.Command, args []string) {
		logger := logging.New()
		logger.Info().Msg("Starting dev server")

		err := viper.BindPFlags(cmd.Flags())
		if err != nil {
			logger.Fatal().Err(err).Msg("failed to bind flags")
		}

		viper.AutomaticEnv()
		viper.SetEnvPrefix("DURABLE")
		viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

		port := viper.GetString("port")

		d, err := durable.New(durable.Config{Logger: logger})
		if err != nil {
			logger.Fatal().Err(err).Send()
		}
		go func() {
			err := d.Run()
			if err != nil {
				logger.Fatal().Err(err).Send()
			}
		}()

		srv, err := server.New(server.Config{
			Durable: d,
			Logger:  logger,
		})
		if err != nil {
			logger.Fatal().Err(err).Send()

		}
		err = srv.Listen(fmt.Sprintf(":%s", port))
		if err != nil {
			logger.Fatal().Err(err).Send()

		}

	},
}

func init() {
	rootCmd.AddCommand(devCmd)
	devCmd.Flags().String("port", "8080", "The port where the dev server listens")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// devCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// devCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
