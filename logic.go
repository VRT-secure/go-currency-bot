package main

import (
	"context"
	"fmt"
	"log"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/looplab/fsm"
)

var mainMenuKeyboard = tgbotapi.NewReplyKeyboard(
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
		ChatID:   chatID,
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

func convertFiatCurrency(fist_fiatCurrency, second_fiatCurrency FiatCurrency, amount int) (string, error) {
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

func mainMenu(text string, keyboard tgbotapi.ReplyKeyboardMarkup) (tgbotapi.ReplyKeyboardMarkup, string) {
	return keyboard, text
}
