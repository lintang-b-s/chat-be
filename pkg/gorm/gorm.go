package gorm

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Gorm struct {
	Pool *gorm.DB
}

func NewGorm() (*Gorm, error) {
	dsn := "host=localhost user=user password=pass dbname=chat port=5431 sslmode=disable TimeZone=Asia/Shanghai"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("gorm - NewGorm - gorm.Open: %w", err)
	}
	gorm := &Gorm{
		Pool: db,
	}

	return gorm, nil
}
