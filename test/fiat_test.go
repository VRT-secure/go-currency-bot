package test

import (
	"testing"

	"kakafoni/database"
	"kakafoni/fiat_currency"
	"kakafoni/metal_price/makhachkala"
	"kakafoni/metal_price/russia"

	. "github.com/smartystreets/goconvey/convey"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestSqlite3(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	Convey("Creating database", t, func() {
		So(err, ShouldBeNil)
	})

	Convey("Create tables", t, func() {
		err := database.CreateTable(db, &fiat_currency.FiatCurrency{})
		So(err, ShouldBeNil)
		err = database.CreateTable(db, &makhachkala.GoldPriceMakhachkala{})
		So(err, ShouldBeNil)
		err = database.CreateTable(db, &russia.MetalPrices{})
		So(err, ShouldBeNil)
	})

	Convey("Parse into tables", t, func() {
		err := fiat_currency.ParseJsonIntoTable(db, fiat_currency.URL_TO_JSON_FIAT)
		So(err, ShouldBeNil)
		err = makhachkala.ParseGoldPriseMakhachkala(db)
		So(err, ShouldBeNil)
		err = russia.ParseCbrMetalPrice(db)
		So(err, ShouldBeNil)
	})

	Convey("Select from tables", t, func() {
		fiatCurrencySlice := fiat_currency.CharCodes(db)
		So(fiatCurrencySlice, ShouldNotBeEmpty)

		_, err := fiat_currency.SelectFromTable(db, fiatCurrencySlice[0].CharCode)
		So(err, ShouldBeNil)

		_, event := fiat_currency.HandleChoice(fiatCurrencySlice,
			fiatCurrencySlice[0].CharCode,
			"pupok",
		)
		So(len(event), ShouldNotEqual, 0)

		_, err = makhachkala.HandleChoice(db)
		So(err, ShouldBeNil)

		_, err = russia.HandleChoice(db)
		So(err, ShouldBeNil)
	})

	Convey("Delete tables", t, func() {
		err := database.DropTable(db, &fiat_currency.FiatCurrency{})
		So(err, ShouldBeNil)
		err = database.DropTable(db, &makhachkala.GoldPriceMakhachkala{})
		So(err, ShouldBeNil)
		err = database.DropTable(db, &russia.MetalPrices{})
		So(err, ShouldBeNil)
	})
}
