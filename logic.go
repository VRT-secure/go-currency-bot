package main

import (
	"log"
	"fmt"
	"strconv"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/thoas/go-funk"
	"github.com/looplab/fsm"
	"context"
)



var choiceOperationKeyboard = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("Узнать курс валюты"),
		tgbotapi.NewKeyboardButton("Калькулятор валют"),
	),
)

type UserFSM struct {
	ChatID   int64
	FSM      *fsm.FSM
	UserData map[string]string
}

func NewUserFSM(chatID int64) *UserFSM {
	u := &UserFSM{
		ChatID: chatID,
		UserData: make(map[string]string),
	}

	u.FSM = fsm.NewFSM(
		"init",
		fsm.Events{
			{Name: "start", Src: []string{"init", 
			"choiceOperation", "choiceOneCurrency", 
			"choiceFirstCurrency", "choiceSecondCurrency", 
			"choiceAmount"}, Dst: "choiceOperation"},
			{Name: "courseCurrency", Src: []string{"choiceOperation"}, Dst: "choiceOneCurrency"},
			{Name: "firstCyrrency", Src: []string{"choiceOperation"}, Dst: "choiceFirstCurrency"},
			{Name: "secondCyrrency", Src: []string{"choiceFirstCurrency"}, Dst: "choiceSecondCurrency"},
			{Name: "amount", Src: []string{"choiceSecondCurrency"}, Dst: "choiceAmount"},
		},
		fsm.Callbacks{
			"enter_state": func(ctx context.Context, e *fsm.Event) {
				u.enterState(e)
			},
		},
	)

	return u
}

// для вывода на экран о смене состояния у пользователя
func (u *UserFSM) enterState(e *fsm.Event) {
	fmt.Printf("User %d entered state %s\n", u.ChatID, e.Dst)
}

func (u *UserFSM) changeEvent(event string) {
	if event != "" {
		err := u.FSM.Event(context.Background(), event)
		if err != nil {
			log.Printf("FSM error: %v", err)
		}
	}
}



func currencyCharCodes(map_currencyes map[string][]string) []string {
	keys := make([]string, 0, len(map_currencyes))
	for key := range map_currencyes {
		keys = append(keys, key)
	}
	return keys
}

func currencyInfo(cyrrency_data []string) string {
	answer := ""
	labels_str := []string {"Абревиатура", "Номинал", "Название", "Цена в рублях на сегодня", "Прошлая цена"}
	for i := 0; i < len(cyrrency_data); i++ {
		answer += labels_str[i] + ": " + cyrrency_data[i] + "\n"
	}
	return answer
}

func handleCurrencyChoice(charCodes []string, charCode string, map_currencyes map[string][]string, nextEvent string) (string, string) {
	if funk.Contains(charCodes, charCode) {
		return currencyInfo(map_currencyes[charCode]), nextEvent
	}
	return "Данная валюта не найдена, отправьте валюту снова или отмените операцию /cancel", ""
}

func convertCurrency(first_currency_code, second_currency_code string, map_currencyes map[string][]string, amount int) (string, error) {
	value_fist_currency, err := strconv.ParseFloat(map_currencyes[first_currency_code][3], 32)
	if err != nil {
		return "", err
	}

	value_second_currency, err := strconv.ParseFloat(map_currencyes[second_currency_code][3], 32)
	if err != nil {
		return "", err
	}

	nominal_fist_currency, err := strconv.Atoi(map_currencyes[first_currency_code][1])
	if err != nil {
		return "", err
	}

	nominal_second_currency, err := strconv.Atoi(map_currencyes[second_currency_code][1])
	if err != nil {
		return "", err
	}

	tmp := value_fist_currency / (value_second_currency / float64(nominal_second_currency)) / float64(nominal_fist_currency) * float64(amount)
	answer := fmt.Sprintf("%.4f", tmp)
	return answer, err
}

func charCodesKeyboard(charCodes []string) tgbotapi.ReplyKeyboardMarkup {
	var rows [][]tgbotapi.KeyboardButton

	for i := 0; i < len(charCodes); i += 3 {
		row := []tgbotapi.KeyboardButton{}

		for j := 0; j < 3 && i+j < len(charCodes); j++ {
			row = append(row, tgbotapi.NewKeyboardButton(charCodes[i+j]))
		}

		rows = append(rows, row)
	}

	return tgbotapi.NewReplyKeyboard(rows...)
}
