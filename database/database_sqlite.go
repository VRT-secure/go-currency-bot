package database

import (
	"gorm.io/gorm"
)

func CreateTable(db *gorm.DB, table interface{}) error {
	err := db.AutoMigrate(table)
	return err
}

func InsertIntoDB(db *gorm.DB, record interface{}) error {
	result := db.Create(record)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func DropTable(db *gorm.DB, table interface{}) error {
	err := db.Migrator().DropTable(table)
	return err
}
