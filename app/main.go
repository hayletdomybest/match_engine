package main

import (
	"log"
	"match_engine/app/cmd"
)

func main() {
	rootCmd := cmd.NewRootCmd()
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
