package server

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewServerCmd() *cobra.Command {
	return &cobra.Command{
		Use: "server",

		RunE: func(cmd *cobra.Command, args []string) error {
			//TODO
			fmt.Println("Implement run server")
			return nil
		},
	}
}
