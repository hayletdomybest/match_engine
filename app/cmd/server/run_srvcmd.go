package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"match_engine/app/cmd/common"
	"os"
	"path"

	"github.com/spf13/cobra"
)

func runCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "run",
		Short: "run server",
		RunE: func(cmd *cobra.Command, args []string) error {
			home := cmd.Context().Value(common.HomeContextKey).(string)

			configPath := path.Join(home, common.DefaultConfigName)

			conf, err := readConfig(configPath)
			if err != nil {
				return err
			}

			if len(conf.DataDir) == 0 {
				conf.DataDir = fmt.Sprintf("%s-%d", common.DefaultDataDir, conf.NodeID)
			}
			ctx := &common.AppContext{
				AppConfig: *conf,
				Home:      home,
			}
			container := initContainer(ctx)
			if err := sealed(container); err != nil {
				return err
			}

			srv, err := getFromContainer(container, &Server{})
			if err != nil {
				return err
			}

			if err := srv.Run(); err != nil {
				return err
			}
			return nil
		},
	}
}

func readConfig(configPath string) (*common.AppConfig, error) {

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
	var config common.AppConfig
	if err := json.Unmarshal(bz, &config); err != nil {
		return nil, err
	}

	return &config, nil
}
