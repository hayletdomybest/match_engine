package common

import "context"

type AppContext struct {
	AppConfig
	Home    string `json:"-"`
	Context context.Context
}
