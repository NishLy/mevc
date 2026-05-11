package room

type RoomSettingRequest struct {
	AutoJoin         *bool   `json:"auto_join,omitempty"`
	AllowGuests      *bool   `json:"allow_guests,omitempty"`
	AllowRecording   *bool   `json:"allow_recording,omitempty"`
	AllowChat        *bool   `json:"allow_chat,omitempty"`
	AllowScreenShare *bool   `json:"allow_screen_share,omitempty"`
	Capacity         *uint   `json:"capacity,omitempty"`
	Location         *string `json:"location,omitempty"`
}

type CreateRoomRequest struct {
	ID          *uint64             `json:"id,omitempty"` // Optional for updates
	Name        string              `json:"name" validate:"required"`
	Description string              `json:"description"`
	Pin         *string             `json:"pin,omitempty" validate:"omitempty,len=4,alphanum"`
	Settings    *RoomSettingRequest `json:"settings" validate:"required"`
}
