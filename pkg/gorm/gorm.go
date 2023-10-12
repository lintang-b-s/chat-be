package gorm

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Gorm struct {
	Pool *gorm.DB
}

func NewGorm(username string, password string) (*Gorm, error) {
	dsn := "host=localhost user=" + username + " password=" + password + " dbname=chat port=5432 sslmode=disable TimeZone=Asia/Shanghai"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("gorm - NewGorm - gorm.Open: %w", err)
	}
	gorm := &Gorm{
		Pool: db,
	}

	return gorm, nil
}
