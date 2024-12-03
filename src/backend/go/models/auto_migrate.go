package models

import (
	"gorm.io/gorm"
)

// AutoMigrateAll will migrate all registered models
func AutoMigrateAll(db *gorm.DB) {
	db.AutoMigrate(
		&Album{},
		&Song{},
	)
}
