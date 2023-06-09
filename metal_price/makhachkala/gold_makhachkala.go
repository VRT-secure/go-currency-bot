package makhachkala

import (
	"fmt"
	"kakafoni/database"
	"kakafoni/metal_price"
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

func ParseGoldPriseMakhachkala(db *gorm.DB) error {
	c := colly.NewCollector()

	var err error = nil
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
				err = insertRecordIntoTable(db, answer[0], answer[1], answer[2])
				if err != nil {
					log.Printf("Ошибка встаски цены золота в БД: %s", err)
					return
				}
			}
		})
	}

	// Начать скрапинг
	for _, url := range metal_price.Gold_url_with_probes_mackhachkala {
		c.Visit(url)
	}

	return err
}

func insertRecordIntoTable(db *gorm.DB, goldContent, priceFrom, priceUpTo string) error {
	currency := &GoldPriceMakhachkala{GoldContent: goldContent, PriceFrom: priceFrom, PriceUpTo: priceUpTo}
	err := database.InsertIntoDB(db, currency)
	return err
}

func selectFromTable(db *gorm.DB) ([]GoldPriceMakhachkala, error) {
	var goldPriceMakhachkala []GoldPriceMakhachkala
	result := db.Order("id DESC").Limit(8).Find(&goldPriceMakhachkala)
	if result.Error != nil {
		log.Printf("Ошибка поиска записей цены золота: %v", result.Error)
	}
	return goldPriceMakhachkala, result.Error
}

func IsTableEmpty(db *gorm.DB) bool {
	var count int64
	db.Model(&GoldPriceMakhachkala{}).Count(&count)
	return count == 0
}

func HandleChoice(db *gorm.DB) (string, error) {
	goldPriceSlice, err := selectFromTable(db)
	if err != nil {
		return "Ошибка операции, отправьте команду снова или отмените операцию /cancel", err
	}
	answer := fmt.Sprintf("Дата полученных данных: %s\n\n", goldPriceSlice[0].CreatedAt.Format("2006-01-02"))
	for _, gold := range goldPriceSlice {
		answer += fmt.Sprintf("Проба %s:\n", gold.GoldContent)
		answer += fmt.Sprintf("Цена от %s\n", gold.PriceFrom)
		answer += fmt.Sprintf("Цена до %s\n\n", gold.PriceUpTo)
	}
	return answer, err
}
