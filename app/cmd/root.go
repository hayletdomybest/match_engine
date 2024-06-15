package cmd

import (
	"context"
	"fmt"
	"match_engine/app/cmd/common"
	"match_engine/app/cmd/server"

	"github.com/spf13/cobra"
)

func NewRootCmd() *cobra.Command {
	var home string

	rootCmd := &cobra.Command{
		Use:   fmt.Sprintf("%sd", common.AppName),
		Short: "matching engine app",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			cmd.SetOut(cmd.OutOrStdout())
			cmd.SetErr(cmd.ErrOrStderr())

			ctx := context.WithValue(context.Background(), common.GlobalContextKey, &common.GlobalContext{
				Home: home,
			})

			cmd.SetContext(ctx)

			return nil
		},
	}

	// Define the config path flag
	rootCmd.PersistentFlags().StringVar(&home, common.HomeFlagName, "", "home path")
	rootCmd.MarkPersistentFlagRequired("home")

	initCmd(rootCmd)
	return rootCmd
}

func initCmd(rootCmd *cobra.Command) {
	rootCmd.AddCommand(
		server.NewServerCmd(),
	)
}
