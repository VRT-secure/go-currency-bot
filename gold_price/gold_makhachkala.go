package gold_price

import (
	"fmt"
	"kakafoni/database"
	"log"
	"strings"

	"github.com/gocolly/colly"
	"gorm.io/gorm"
)

type GoldPriceMakhachkala struct {
	gorm.Model
	GoldContent string
	PriceFrom   string
	PriceUpTo   string
}

func Parse_gold_prise_makhachkala(db *gorm.DB) error {
	c := colly.NewCollector()

	answer := make([]string, 3)
	for i := 2; i <= 4; i++ {
		i := i // Создаем новую переменную i для каждой итерации, ибо у нас замыкание.
		c.OnHTML(fmt.Sprintf("td:nth-child(%d)", i), func(e *colly.HTMLElement) {
			text := strings.TrimSpace(e.Text)
			switch i {
			case 2:
				answer[0] = text
			case 3:
				answer[1] = text
			case 4:
				answer[2] = text
			}
		})
	}

	// Начать скрапинг
	for _, url := range gold_url_with_probes_mackhachkala {
		c.Visit(url)
	}
	err := insertRecordIntoTable(db, answer[0], answer[1], answer[2])
	return err
}

func insertRecordIntoTable(db *gorm.DB, goldContent, priceFrom, priceUpTo string) error {
	currency := &GoldPriceMakhachkala{GoldContent: goldContent, PriceFrom: priceFrom, PriceUpTo: priceUpTo}
	err := database.InsertIntoDB(db, currency)
	return err
}

func SelectFromTable(db *gorm.DB, userText string) (GoldPriceMakhachkala, error) {
	var goldPrice GoldPriceMakhachkala
	result := db.Last(&goldPrice)
	if result.Error != nil {
		log.Printf("Ошибка чтения фиатной валюты из БД: %v", result.Error)
		return GoldPriceMakhachkala{}, result.Error
	}
	return goldPrice, nil
}

func IsTableEmpty(db *gorm.DB) bool {
	var count int64
	db.Model(&GoldPriceMakhachkala{}).Count(&count)
	return count == 0
}
