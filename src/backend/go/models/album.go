package models

type Album struct {
	ID          uint   `gorm:"primaryKey"`
	Name        string `gorm:"not null"`
	PicFilePath string `gorm:"not null"`

	Songs []Song
}
