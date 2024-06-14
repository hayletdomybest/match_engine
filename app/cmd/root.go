package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"match_engine/app/cmd/server"
	"os"

	"github.com/spf13/cobra"
)

func NewRootCmd() *cobra.Command {
	var configPath string

	rootCmd := &cobra.Command{
		Use:   fmt.Sprintf("%sd", AppName),
		Short: "matching engine app",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			cmd.SetOut(cmd.OutOrStdout())
			cmd.SetErr(cmd.ErrOrStderr())

			appCtx, err := initConfig(configPath)
			if err != nil {
				return err
			}

			ctx := context.WithValue(context.Background(), ServerContext, appCtx)
			cmd.SetContext(ctx)

			<-appCtx.sealedch

			return nil
		},
	}

	// Define the config path flag
	rootCmd.PersistentFlags().StringVar(&configPath, "config", "", "Path to the configuration file")
	rootCmd.MarkPersistentFlagRequired("config")

	initCmd(rootCmd)
	return rootCmd
}

func initCmd(rootCmd *cobra.Command) {
	rootCmd.AddCommand(
		server.NewServerCmd(),
	)
}

func initConfig(configPath string) (*AppContext, error) {

	// Validate config path
	if configPath == "" {
		return nil, errors.New("config path is required")
	}

	// Read the config file
	configFile, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer configFile.Close()

	// Read the file content
	bz, err := io.ReadAll(configFile)
	if err != nil {
		return nil, err
	}

	// Unmarshal the JSON content
	var config AppConfig
	if err := json.Unmarshal(bz, &config); err != nil {
		return nil, err
	}

	// Create a new AppContext
	ctx := NewContext(config.NodeID, config.URL)

	for id, url := range config.Peers {
		ctx.AppendPeer(id, url)
	}

	ctx.Sealed()
	return ctx, nil
}
