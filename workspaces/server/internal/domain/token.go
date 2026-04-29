package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Token struct {
	TimestampModel
	ID      uuid.UUID `gorm:"primaryKey;not null" json:"id"`
	Token   string    `gorm:"not null" json:"token"`
	UserID  uuid.UUID `gorm:"not null" json:"user_id"`
	Type    string    `gorm:"not null" json:"type"`
	Expires time.Time `gorm:"not null" json:"expires"`
	User    *User     `gorm:"foreignKey:user_id;references:id" json:"user,omitempty"`
}

func (token *Token) BeforeCreate(_ *gorm.DB) error {
	token.ID = uuid.New()
	return nil
}
