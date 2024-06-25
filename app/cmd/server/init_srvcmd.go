package server

import (
	"encoding/json"
	"errors"
	"match_engine/app/cmd/common"
	"match_engine/utils"
	"os"
	"path"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
)

func initCmd() *cobra.Command {
	var overwrite bool = false
	cmd := &cobra.Command{
		Use:   "init",
		Short: "init server",

		RunE: func(cmd *cobra.Command, args []string) error {
			home := cmd.Context().Value(common.HomeContextKey).(string)
			return initConfig(path.Join(home, common.DefaultConfigName), overwrite)
		},
	}

	cmd.Flags().BoolVarP(&overwrite, "overwrite", "o", false, "overwrite config")
	return cmd
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
	defaultConfig := common.AppConfig{
		Mode:    gin.DebugMode,
		ApiPort: 3000,
		NodeID:  1,
		NodeUrl: "http://127.0.0.1:8081",
		Peers: map[uint64]string{
			1: "http://127.0.0.1:8081",
		},
		Join:          false,
		EtchEndpoints: make([]string, 0),
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
