package test

import (
	"testing"

	"kakafoni/database"
	"kakafoni/fiat_currency"

	. "github.com/smartystreets/goconvey/convey"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestSqlite3Fiat(t *testing.T) {
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
		fiatCurrencySlice := fiat_currency.CharCodes(db)
		So(fiatCurrencySlice, ShouldNotBeEmpty)

		_, err := fiat_currency.SelectFromTable(db, fiatCurrencySlice[0].CharCode)
		So(err, ShouldBeNil)

		_, event := fiat_currency.HandleCurrencyChoice(fiatCurrencySlice,
			fiatCurrencySlice[0].CharCode,
			"pupok",
		)
		So(len(event), ShouldNotEqual, 0)
	})

	Convey("Delete FiatCurrency table", t, func() {
		err := database.DropTable(db, &fiat_currency.FiatCurrency{})
		So(err, ShouldBeNil)
	})
}
