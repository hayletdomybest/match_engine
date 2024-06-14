package cmd

import (
	"errors"
	"fmt"
)

// AppConfig represents the configuration for the application
type AppConfig struct {
	NodeID uint64            `json:"node_id"`
	URL    string            `json:"url"`
	Peers  map[uint64]string `json:"peers"`
}

// Validate checks if the configuration is valid
func (config *AppConfig) Validate() error {
	if config.NodeID == 0 {
		return errors.New("node ID is required")
	}
	if config.URL == "" {
		return errors.New("URL is required")
	}
	if len(config.Peers) == 0 {
		return errors.New("at least one peer is required")
	}
	for id, peer := range config.Peers {
		if id == 0 || peer == "" {
			return fmt.Errorf("invalid peer entry: %d -> %s", id, peer)
		}
	}
	return nil
}
