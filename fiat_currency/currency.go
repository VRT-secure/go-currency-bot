package fiat_currency

import (
	"io"
	"log"
	"net/http"
	"time"

	"github.com/tidwall/gjson"
	"gorm.io/gorm"
	"github.com/thoas/go-funk"
	"kakafoni/database"
)


type FiatCurrency struct {
	gorm.Model
	CharCode string
	Nominal  string
	Name     string
	Value    string
	Previous string
}

func parseJsonIntoTable(db *gorm.DB, url_to_json string) {
	// Создаем HTTP-запрос для скачивания файла
	resp, err := http.Get(url_to_json)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	// Получаем массив байтов
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	json := string(body)

	gjson.Get(json, "Valute").ForEach(func(key, value gjson.Result) bool {
		err := insertFiatRecordIntoTable(db,
			value.Get("CharCode").String(),
			value.Get("Nominal").String(),
			value.Get("Name").String(),
			value.Get("Value").String(),
			value.Get("Previous").String(),
		)
		if err != nil {
			log.Print("Не удалось записать данные фиатных валют в БД.", err)
		}
		return true // продолжить итерацию
	})
}

func insertFiatRecordIntoTable(db *gorm.DB, charCode, nominal, name, value, previous string) error {
	currency := &FiatCurrency{CharCode: charCode, Nominal: nominal, Name: name, Value: value, Previous: previous}
	err := database.InsertIntoDB(db, currency)
	return err
}

func selectFiatFromTable(db *gorm.DB, charCode string) (FiatCurrency, error) {
	var fiatCurrency FiatCurrency
	result := db.Last(&fiatCurrency, "char_code = ?", charCode)
	if result.Error != nil {
		log.Printf("Ошибка чтения фиатной валюты из БД: %v", result.Error)
		return FiatCurrency{}, result.Error
	}
	return fiatCurrency, nil
}

func isFiatTableEmpty(db *gorm.DB) bool {
	var count int64
	db.Model(&FiatCurrency{}).Count(&count)
	return count == 0
}

func fiatCharCodes(db *gorm.DB) []string {
	var fiatCurrency []FiatCurrency
	location, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		log.Printf("Ошибка установки московского часового пояса для поиска фиатных валют: %v", err)
	}
	now := time.Now().In(location)
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, location)
	endOfDay := startOfDay.Add(24 * time.Hour)
	result := db.Where("created_at >= ? AND created_at < ?", startOfDay, endOfDay).Find(&fiatCurrency)
	if result.Error != nil {
		log.Printf("Ошибка поиска записей фиатных валют в БД по текущей дате: %v", result.Error)
	}
	
	var answer_str []string
	for _, fiat := range fiatCurrency {
		answer_str = append(answer_str, fiat.CharCode)
	}
	return answer_str
}

func handleFiatCurrencyChoice(charCodes []string, charCode string, fiatCurrency FiatCurrency, nextEvent string) (string, string) {
	if funk.Contains(charCodes, charCode) {
		answer := "Абревиатура: " + fiatCurrency.CharCode +
			"\nНоминал: " + fiatCurrency.Nominal +
			"\nНазвание: " + fiatCurrency.Name +
			"\nЦена в рублях на сегодня: " + fiatCurrency.Value +
			"\nПрошлая цена: " + fiatCurrency.Previous
		return answer, nextEvent
	}
	return "Данная валюта не найдена, отправьте валюту снова или отмените операцию /cancel", ""
}
