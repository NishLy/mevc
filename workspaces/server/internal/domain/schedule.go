package domain

import "time"

type Schedule struct {
	TimestampModel
	SoftDeleteModel
	ID      uint64    `json:"id" gorm:"primaryKey"`
	RoomID  uint64    `json:"room_id" gorm:"index;not null"`
	Room    *Room     `gorm:"foreignKey:RoomID;references:ID" json:"room,omitempty"`
	Start   time.Time `json:"start" gorm:"not null"`
	End     time.Time `json:"end" gorm:"not null"`
	Pattern string    `json:"pattern" gorm:"not null"` // e.g., "daily", "weekly", "monthly"
}
