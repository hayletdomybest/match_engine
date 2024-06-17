package common

type AppContext struct {
	AppConfig
	Home string `json:"-"`
}
