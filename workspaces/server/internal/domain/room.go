package domain

import (
	"fmt"

	pkg "github.com/NishLy/go-fiber-boilerplate/pkg/hash"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Room struct {
	TimestampModel
	SoftDeleteModel
	ID               uint64     `json:"id" gorm:"primaryKey"`
	Code             string     `json:"code" gorm:"uniqueIndex;size:8;not null"` // Unique 8-character code
	Name             string     `json:"name" gorm:"not null"`
	Description      string     `json:"description"`
	HostID           uuid.UUID  `json:"host_id" gorm:"not null"`
	Url              string     `json:"url" gorm:"not null"`
	Host             *User      `gorm:"foreignKey:HostID;references:ID" json:"host,omitempty"`
	Pin              *string    `json:"-"` // Exclude hashed pin from JSON responses
	Schedules        []Schedule `gorm:"foreignKey:RoomID;references:ID" json:"schedules,omitempty"`
	AutoJoin         bool       `json:"auto_join" gorm:"default:true"`
	AllowGuests      bool       `json:"allow_guests" gorm:"default:false"`
	AllowRecording   bool       `json:"allow_recording" gorm:"default:false"`
	AllowChat        bool       `json:"allow_chat" gorm:"default:true"`
	AllowScreenShare bool       `json:"allow_screen_share" gorm:"default:true"`
	Capacity         uint       `json:"capacity" gorm:"default:10"`
	Location         string     `json:"location"`
}

func (room *Room) BeforeCreate(_ *gorm.DB) error {
	if room.Pin != nil {
		hashedPin, err := pkg.HashPassword(*room.Pin)

		if err != nil {
			return err
		}

		room.Pin = &hashedPin
	}

	room.Code = pkg.GenerateUniqueCode(8)
	room.Url = fmt.Sprintf("/room/%s", room.Code)

	return nil
}
