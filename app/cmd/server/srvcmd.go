package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"match_engine/app/cmd/common"
	"match_engine/utils"
	"os"
	"path"
	"path/filepath"

	"github.com/spf13/cobra"
)

func NewServerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "server",
	}

	cmd.AddCommand(
		initCmd(),
		runCmd(),
	)

	cmd.AddCommand()
	return cmd
}

func initCmd() *cobra.Command {
	var overwrite bool = false
	cmd := &cobra.Command{
		Use:   "init",
		Short: "init server",

		RunE: func(cmd *cobra.Command, args []string) error {
			globalCtx := cmd.Context().Value(common.GlobalContextKey).(*common.GlobalContext)
			return initConfig(path.Join(globalCtx.Home, common.DefaultConfigName), overwrite)
		},
	}

	cmd.Flags().BoolVarP(&overwrite, "overwrite", "o", false, "overwrite config")
	return cmd
}

func runCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "run",
		Short: "run server",
		RunE: func(cmd *cobra.Command, args []string) error {
			globalCtx := cmd.Context().Value(common.GlobalContextKey).(*common.GlobalContext)

			configPath := path.Join(globalCtx.Home, common.DefaultConfigName)

			conf, err := readConfig(configPath)
			if err != nil {
				return err
			}

			server := NewServer(conf.URL)
			server.inject(func() *common.GlobalContext {
				return globalCtx
			})
			server.RegisterController()
			server.RegisterRepository()

			fmt.Printf("Run server %s", conf.URL)
			if err := server.Run(); err != nil {
				return err
			}
			return nil
		},
	}
}

func initConfig(configPath string, overwrite bool) error {
	// Validate config path
	if configPath == "" {
		return errors.New("config path is required")
	}

	// Read the config file
	if !overwrite {
		configFile, err := os.Open(configPath)
		if err == nil {
			defer configFile.Close()
			return nil
		}
	}

	configDir := filepath.Dir(configPath)
	if err := utils.MkdirAll(configDir); err != nil {
		return err
	}

	// If the file does not exist or cannot be read, create a default config
	defaultConfig := ServerConfig{
		NodeID: 1,
		URL:    "127.0.0.1:8081",
		Peers: map[uint64]string{
			1: "127.0.0.1:8081",
		},
	}

	// Create the file and write the default config
	bz, err := json.MarshalIndent(&defaultConfig, "", "  ")
	if err != nil {
		return err
	}

	err = os.WriteFile(configPath, bz, 0644)
	if err != nil {
		return err
	}

	return nil
}

func readConfig(configPath string) (*ServerContext, error) {

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
	var config ServerConfig
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
