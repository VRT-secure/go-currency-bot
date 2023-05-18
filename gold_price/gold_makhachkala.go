package gold_price

import (
	"fmt"
	"kakafoni/database"
	"log"
	"strings"
	"time"

	"github.com/gocolly/colly"
	"gorm.io/gorm"
)

type GoldPriceMakhachkala struct {
	gorm.Model
	GoldContent string
	PriceFrom   string
	PriceUpTo   string
}

// TODO выяснить почему в таблицу записывается только последняя проба~
func ParseGoldPriseMakhachkala(db *gorm.DB) error {
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

	var err error = nil
	for i := 0; i < len(answer); i += 3 {
		err = insertRecordIntoTable(db, answer[i], answer[i+1], answer[i+2])
		if err != nil {
			log.Printf("Ошибка встаски цены золота в БД: %s", err)
			return err
		}
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
	location, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		log.Printf("Ошибка установки московского часового пояса для поиска: %v", err)
	}
	now := time.Now().In(location)
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, location)
	endOfDay := startOfDay.Add(24 * time.Hour)
	result := db.Where("created_at >= ? AND created_at < ?", startOfDay, endOfDay).Find(&goldPriceMakhachkala)
	if result.Error != nil {
		log.Printf("Ошибка поиска записей в БД по текущей дате: %v", result.Error)
	}

	return goldPriceMakhachkala, result.Error
}

func IsTableEmpty(db *gorm.DB) bool {
	var count int64
	db.Model(&GoldPriceMakhachkala{}).Count(&count)
	return count == 0
}

func HandleChoice(db *gorm.DB, nextEvent string) (string, string) {
	goldPriceSlice, err := selectFromTable(db)
	if err != nil {
		return "Ошибка операции, отправьте команду снова или отмените операцию /cancel", ""
	}
	answer := ""
	for _, gold := range goldPriceSlice {
		answer += fmt.Sprintf("Проба %s", gold.GoldContent)
		answer += fmt.Sprintf("Цена от %s", gold.PriceFrom)
		answer += fmt.Sprintf("Цена до %s", gold.PriceUpTo)
	}
	return answer, nextEvent
}
