package main

import (
	"io"
	"log"
	"net/http"

	"github.com/tidwall/gjson"
	"gorm.io/gorm"
)

const URL = "https://www.cbr-xml-daily.ru/daily_json.js"

type FiatCurrency struct {
	gorm.Model
	CharCode string
	Nominal  string
	Name     string
	Value    string
	Previous string
}


func parseJsonIntoTable(url_to_json string) map[string][]string { // TODO переделать функцию для заполнения таблицы
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

	map_currency := make(map[string][]string)
	gjson.Get(json, "Valute").ForEach(func(key, value gjson.Result) bool {
		map_currency[value.Get("CharCode").String()] = []string{
			value.Get("CharCode").String(),
			value.Get("Nominal").String(),
			value.Get("Name").String(),
			value.Get("Value").String(),
			value.Get("Previous").String(),
		}
		return true // продолжить итерацию
	})

	return map_currency
}

func insertFiatRecordIntoTable(db *gorm.DB, charCode, nominal, name, value, previous string) {
	currency := &FiatCurrency{CharCode: charCode, Nominal: nominal, Name: name, Value: value, Previous: previous}
	err := InsertIntoDB(db, currency)
	if err != nil {
		log.Print("Не удалось записать данные фиатных валют в БД.", err)
	}
}

func selectFiatFromTable(db *gorm.DB, charCode string) (FiatCurrency, error) {
	var fiatCurrency FiatCurrency
	result := db.Last(&fiatCurrency, "charCode = ?", charCode)
	if result.Error != nil {
		return FiatCurrency{}, result.Error
	}
	return fiatCurrency, nil
}

func isFiatTableEmpty(db *gorm.DB) bool {
	var count int64
	db.Model(&FiatCurrency{}).Count(&count)
	if count == 0 {
		return true
	}
	return false
}
