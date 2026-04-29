package domain

import (
	"time"

	"gorm.io/gorm"
)

type SoftDeleteModel struct {
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

type TimestampModel struct {
	CreatedAt time.Time `gorm:"autoCreateTime:milli" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoCreateTime:milli;autoUpdateTime:milli" json:"updated_at"`
}
