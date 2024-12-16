package models

type Song struct {
	ID                uint   `gorm:"primaryKey"`
	Name              string `gorm:"not null"`
	AudioFilePath     string `gorm:"not null"`
	AudioFilePathMidi string `gorm:"not null"`
	MidiJSON          string `gorm:"not null"`

	AlbumID *uint
	Album   Album
}
