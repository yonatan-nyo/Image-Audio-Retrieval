package models

type Song struct {
	ID            uint `gorm:"primaryKey"`
	Name          *string
	AudioFilePath string `gorm:"not null"`

	AlbumID *uint
	Album   Album
}
