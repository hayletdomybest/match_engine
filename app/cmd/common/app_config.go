package common

type AppConfig struct {
	ApiPort uint64            `json:"api_port"`
	NodeID  uint64            `json:"node_id"`
	URL     string            `json:"url"`
	Peers   map[uint64]string `json:"peers"`
	Join    bool              `json:"join"`
	DataDir string            `json:"data_dir"`
}
