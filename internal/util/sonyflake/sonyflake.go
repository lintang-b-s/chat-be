package sonyflake

import (
	"fmt"
	"github.com/sony/sonyflake"
)

type SonyFlake struct {
	sf *sonyflake.Sonyflake
}

func NewSonyFlake() *SonyFlake {
	var st sonyflake.Settings
	sf := sonyflake.NewSonyflake(st)
	return &SonyFlake{sf}
}

func (sf *SonyFlake) GenerateId() (uint64, error) {
	id, err := sf.sf.NextID()
	if err != nil {
		return 0, fmt.Errorf("SonyFlake - GenearateId - sf.NextId()", err)
	}
	return id, nil
}
