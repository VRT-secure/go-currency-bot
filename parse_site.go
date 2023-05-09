package main

import (
	"io"
	"net/http"
	"github.com/tidwall/gjson"
)

const URL = "https://www.cbr-xml-daily.ru/daily_json.js"

func parse_json(url string) map[string][]string {
	// Создаем HTTP-запрос для скачивания файла
	resp, err := http.Get(URL)
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
