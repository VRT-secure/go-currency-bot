package test

import (
	"testing"

	"kakafoni/database"
	"kakafoni/fiat_currency"

	. "github.com/smartystreets/goconvey/convey"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestSqlite3(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	Convey("Creating database file", t, func() {
		So(err, ShouldBeNil)
	})

	Convey("Create FiatCurrency table", t, func() {
		err := database.CreateTable(db, &fiat_currency.FiatCurrency{})
		So(err, ShouldBeNil)
	})

	Convey("Parse json into FiatCurrency table", t, func() {
		err := fiat_currency.ParseJsonIntoTable(db, fiat_currency.URL_TO_JSON_FIAT)
		So(err, ShouldBeNil)
	})

	Convey("Select from FiatCurrency table", t, func() {
		charCodes := fiat_currency.FiatCharCodes(db)
		So(charCodes, ShouldNotBeEmpty)
		_, err := fiat_currency.SelectFiatFromTable(db, charCodes[0])
		So(err, ShouldBeNil)
	})

	Convey("Delete FiatCurrency table", t, func() {
		err := database.DropTable(db, &fiat_currency.FiatCurrency{})
		So(err, ShouldBeNil)
	})
}
