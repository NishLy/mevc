package domain

// Shortener represents a URL shortener entity in the database.
type Shortener struct {
	TimestampModel
	SoftDeleteModel
	ID   uint64 `json:"id" gorm:"primaryKey"`
	Code string `json:"code" gorm:"uniqueIndex;size:16;index;not null"` // Unique 16-character code
	Url  string `json:"url" gorm:"not null"`
}
