package russia

import (
	"fmt"
	"kakafoni/database"
	"kakafoni/metal_price"
	"log"

	"github.com/gocolly/colly"
	"gorm.io/gorm"
)

type MetalPrices struct {
	gorm.Model
	Date      string
	Gold      string
	Silver    string
	Platinum  string
	Palladium string
}

func ParseCbrMetalPrice(db *gorm.DB) error {

	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/80.0.3987.149 Safari/537.36"),
	)
	var err error = nil

	c.OnHTML("table[class='mfd-table'] > tbody > tr", func(e *colly.HTMLElement) {
		metalPrice := MetalPrices{
			Date:      e.ChildText("td:nth-child(1)"),
			Gold:      e.ChildText("td:nth-child(2)"),
			Silver:    e.ChildText("td:nth-child(3)"),
			Platinum:  e.ChildText("td:nth-child(4)"),
			Palladium: e.ChildText("td:nth-child(5)"),
		}
		err = insertRecordIntoTable(db, metalPrice.Date, metalPrice.Gold, metalPrice.Silver, metalPrice.Platinum, metalPrice.Palladium)
		if err != nil {
			log.Printf("Failed to insert record into DB: %v", err)
		}
	})

	c.Visit(metal_price.Url_metal_price_cbr)
	return err
}

func insertRecordIntoTable(db *gorm.DB, date, gold, silver, platinum, palladium string) error {
	metal := &MetalPrices{Date: date, Gold: gold, Silver: silver, Platinum: platinum, Palladium: palladium}
	err := database.InsertIntoDB(db, metal)
	return err
}

func selectFromTable(db *gorm.DB) ([]MetalPrices, error) {
	var MetalPrices []MetalPrices
	result := db.Order("id DESC").Limit(119).Find(&MetalPrices)
	if result.Error != nil {
		log.Printf("Ошибка поиска записей цены золота: %v", result.Error)
	}
	return MetalPrices, result.Error 
}

func IsTableEmpty(db *gorm.DB) bool {
	var count int64
	db.Model(&MetalPrices{}).Count(&count)
	return count == 0
}

func HandleChoice(db *gorm.DB) (string, error) {
	metalPriceSlice, err := selectFromTable(db)
	if err != nil {
		return "Ошибка операции, отправьте команду снова или отмените операцию /cancel", err
	}
	answer := "Дата\t|\tЗолото\t|\tСеребро\t|\tПлатина\t|\tПаладий\n"
	for _, metal := range metalPriceSlice {
		answer += fmt.Sprintf("%s|%s|%s|%s|%s\n", metal.Date, metal.Gold, metal.Silver, metal.Platinum, metal.Palladium)
	}
	return answer, err
}
