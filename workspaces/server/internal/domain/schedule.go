package domain

import (
	"time"

	"gorm.io/datatypes"
)

type Schedule struct {
	TimestampModel
	SoftDeleteModel
	ID      uint64         `json:"id" gorm:"primaryKey"`
	RoomID  uint64         `json:"room_id" gorm:"index;not null"`
	Room    *Room          `gorm:"foreignKey:RoomID;references:ID" json:"room,omitempty"`
	Start   time.Time      `json:"start" gorm:"not null"`
	End     time.Time      `json:"end" gorm:"not null"`
	Pattern datatypes.JSON `json:"pattern" gorm:"not null; type:jsonb"` // e.g., "daily", "weekly", "monthly"
}
