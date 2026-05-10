package domain

import "time"

type Occurance struct {
	TimestampModel
	SoftDeleteModel
	ID          uint64    `json:"id" gorm:"primaryKey"`
	ScheduleID  uint64    `json:"schedule_id" gorm:"index;not null"`
	Schedule    *Schedule `gorm:"foreignKey:ScheduleID;references:ID" json:"schedule,omitempty"`
	Start       time.Time `json:"start" gorm:"not null"` // Unix timestamp
	End         time.Time `json:"end" gorm:"not null"`   // Unix timestamp
	IsCancelled bool      `json:"is_cancelled" gorm:"default:false"`
}
