package test

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"kakafoni/database"
)

func testSqlite3(t *testing.T) {
	db, _ := gorm.Open(sqlite.Open("currencyes.db"), &gorm.Config{})
	database.CreateTable(db, &FiatCurrency{})
	Convey("Given some integer with a starting value", t, func() {

		Convey("When the integer is incremented", func() {

		})
	})
}
