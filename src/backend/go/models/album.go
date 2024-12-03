package models

type Album struct {
	ID   uint `gorm:"primaryKey"`
	Name string

	Songs []Song
}
