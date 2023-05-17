package fiat_currency

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"kakafoni/database"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/tidwall/gjson"
	"gorm.io/gorm"
)

type FiatCurrency struct {
	gorm.Model
	CharCode string
	Nominal  string
	Name     string
	Value    string
	Previous string
}

func ParseJsonIntoTable(db *gorm.DB, url_to_json string) error {
	// Создаем HTTP-запрос для скачивания файла
	resp, err := http.Get(url_to_json)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Получаем массив байтов
	body, err := io.ReadAll(resp.Body)
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
	return err
}

func insertFiatRecordIntoTable(db *gorm.DB, charCode, nominal, name, value, previous string) error {
	currency := &FiatCurrency{CharCode: charCode, Nominal: nominal, Name: name, Value: value, Previous: previous}
	err := database.InsertIntoDB(db, currency)
	return err
}

func SelectFromTable(db *gorm.DB, userText string) (FiatCurrency, error) {
	var fiatCurrency FiatCurrency
	fiatCurrencySlice := CharCodes(db)
	for _, v := range fiatCurrencySlice {
		if strings.Contains(userText, v.CharCode) {
			result := db.Last(&fiatCurrency, "char_code = ?", v.CharCode)
			if result.Error != nil {
				log.Printf("Ошибка чтения фиатной валюты из БД: %v", result.Error)
				return FiatCurrency{}, result.Error
			}
		}
	}
	return fiatCurrency, nil
}

func IsTableEmpty(db *gorm.DB) bool {
	var count int64
	db.Model(&FiatCurrency{}).Count(&count)
	return count == 0
}

func CharCodes(db *gorm.DB) []FiatCurrency {
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

	return fiatCurrency
}

func HandleCurrencyChoice(fiatCurrencySlice []FiatCurrency, userText string, nextEvent string) (string, string) {
	for _, fiat := range fiatCurrencySlice {
		if strings.Contains(userText, fiat.CharCode) {
			answer := "Абревиатура: " + fiat.CharCode +
				"\nНоминал: " + fiat.Nominal +
				"\nНазвание: " + fiat.Name +
				"\nЦена в рублях на сегодня: " + fiat.Value +
				"\nПрошлая цена: " + fiat.Previous
			return answer, nextEvent
		}
	}
	return "Данная валюта не найдена, отправьте валюту снова или отмените операцию /cancel", ""
}

func ConvertCurrency(fist_fiatCurrency, second_fiatCurrency FiatCurrency, amount int) (string, error) {
	value_fist_currency, err := strconv.ParseFloat(fist_fiatCurrency.Value, 32)
	if err != nil {
		return "", err
	}

	value_second_currency, err := strconv.ParseFloat(second_fiatCurrency.Value, 32)
	if err != nil {
		return "", err
	}

	nominal_fist_currency, err := strconv.Atoi(fist_fiatCurrency.Nominal)
	if err != nil {
		return "", err
	}

	nominal_second_currency, err := strconv.Atoi(second_fiatCurrency.Nominal)
	if err != nil {
		return "", err
	}

	tmp := value_fist_currency / (value_second_currency / float64(nominal_second_currency)) / float64(nominal_fist_currency) * float64(amount)
	answer := fmt.Sprintf("%.4f", tmp)
	return answer, err
}

func CharCodesKeyboard(fiatCurrency []FiatCurrency) tgbotapi.ReplyKeyboardMarkup {
	var rows [][]tgbotapi.KeyboardButton

	for i := 0; i < len(fiatCurrency); i++ {
		row := []tgbotapi.KeyboardButton{}

		for j := 0; j < 1 && i+j < len(fiatCurrency); j++ {
			row = append(row, tgbotapi.NewKeyboardButton(fiatCurrency[i+j].CharCode+fmt.Sprintf(" (%s)", fiatCurrency[i+j].Name)))
		}

		rows = append(rows, row)
	}

	return tgbotapi.NewReplyKeyboard(rows...)
}
