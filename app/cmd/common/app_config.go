package common

type AppConfig struct {
	Mode          string            `json:"mode"`
	ApiPort       uint64            `json:"api_port"`
	NodeID        uint64            `json:"node_id"`
	NodeUrl       string            `json:"node_url"`
	Peers         map[uint64]string `json:"peers"`
	Join          bool              `json:"join"`
	DataDir       string            `json:"data_dir"`
	EtchEndpoints []string          `json:"etcd_endpoints"`
}
