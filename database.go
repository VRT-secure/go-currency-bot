package main

import (
	"gorm.io/gorm"
)

func createTable(db *gorm.DB, table interface{}) {
	err := db.AutoMigrate(table)
	if err != nil {
		panic("failed to migrate database")
	}
}

func InsertIntoDB(db *gorm.DB, record interface{}) error {
    result := db.Create(record)
    if result.Error != nil {
        return result.Error
    }
    return nil
}



