package schedule

import "time"

type UpsertScheduleRequest struct {
	ID      uint64                  `json:"id" gorm:"primaryKey"`
	RoomID  uint64                  `json:"room_id" validate:"required"`
	Start   time.Time               `json:"start_at" validate:"required,gt"`
	End     time.Time               `json:"end_at" validate:"required,gtfield=Start"`
	Pattern *PatternScheduleRequest `json:"pattern" validate:"required"`
}

type PatternScheduleRequest struct {
	Frequency string `json:"frequency" validate:"required,oneof=daily weekly monthly"`
	Interval  int    `json:"interval" validate:"required,min=1"`

	// Optional fields based on frequency
	Weekdays []int `json:"weekdays,omitempty" validate:"required_if=Frequency weekly,dive,min=0,max=6"`
	MonthDay *int  `json:"month_day,omitempty" validate:"required_if=Frequency monthly,omitempty,min=1,max=31"`
}
