package utils

import (
	"time"

	"github.com/sony/sonyflake"
)

type IDGenerator interface {
	Generate() (uint64, error)
}

var _ IDGenerator = (*SnowFlake)(nil)

type SnowFlake struct {
	flake *sonyflake.Sonyflake
}

func NewSnowFlake() *SnowFlake {
	settings := sonyflake.Settings{
		StartTime: time.Now().AddDate(-1, 0, 0),
	}
	flake := sonyflake.NewSonyflake(settings)
	return &SnowFlake{
		flake: flake,
	}
}

func (snow *SnowFlake) Generate() (uint64, error) {
	return snow.flake.NextID()
}
