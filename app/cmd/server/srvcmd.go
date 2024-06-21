package server

import (
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
	return cmd
}
