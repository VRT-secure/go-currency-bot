package main

import (
	"log"
	"os"
	"strconv"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	token := os.Getenv("TELEGRAM_APITOKEN")
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60

	updates := bot.GetUpdatesChan(updateConfig)

	userFSMs := make(map[int64]*UserFSM)

	map_currencyes := parse_json(URL)
	charCodes := currencyCharCodes(map_currencyes)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		chatID := update.Message.Chat.ID

		userFSM, ok := userFSMs[chatID]
		if !ok {
			userFSM = NewUserFSM(chatID)
			userFSMs[chatID] = userFSM
		}

		var event string

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
		switch update.Message.Command() {
		case "start":
			event = "start"
			msg.Text = "Выберите операцию"
			msg.ReplyMarkup = choiceOperationKeyboard
		case "help":
			msg.Text = "bot show currency course"
			bot.Send(msg)
			continue
		case "cancel":
			event = "start"
			userFSM.changeEvent(event)
			msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
			bot.Send(msg)
			msg.Text = "Операция отменена"
			msg.ReplyMarkup = choiceOperationKeyboard
		default:
			event = ""
		}

		if userFSM.FSM.Current() == "choiceOperation" {
			choice := update.Message.Text
			if choice == "Узнать курс валюты" {
				msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
				bot.Send(msg)
				msg.ReplyMarkup = charCodesKeyboard(charCodes)
				event = "courseCurrency"
				msg.Text = "Выберите валюту"
			} else if choice == "Калькулятор валют" {
				msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
				bot.Send(msg)
				msg.ReplyMarkup = charCodesKeyboard(charCodes)
				event = "firstCyrrency"
				msg.Text = "Выберите первую валюту"
			} else {
				msg.Text = "Ошибка, такой операции не существует, попробуйте снова"
			}
		} else if userFSM.FSM.Current() == "choiceOneCurrency" {
			charCode := update.Message.Text
			msg.Text, event = handleCurrencyChoice(charCodes, charCode, map_currencyes, "start")
			bot.Send(msg)
			msg.ReplyMarkup = choiceOperationKeyboard
			msg.Text = "Выберите операцию"

		} else if userFSM.FSM.Current() == "choiceFirstCurrency" {
			charCode := update.Message.Text
			msg.Text, event = handleCurrencyChoice(charCodes, charCode, map_currencyes, "secondCyrrency")
			bot.Send(msg)
			userFSM.UserData["firstCurrencyCode"] = charCode
			msg.Text = "Выберите вторую валюту"

		} else if userFSM.FSM.Current() == "choiceSecondCurrency" {
			charCode := update.Message.Text
			msg.Text, event = handleCurrencyChoice(charCodes, charCode, map_currencyes, "amount")
			userFSM.UserData["secondCurrencyCode"] = charCode
			msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
			msg.Text = "Введите количество целым числом"

		} else if userFSM.FSM.Current() == "choiceAmount" {
			amount, err := strconv.Atoi(update.Message.Text)
			if err != nil {
				msg.Text, event = "Ошибка, введите целое число или отменити операцию /cancel", ""
			}else {
				answer, err := convertCurrency(
					userFSM.UserData["firstCurrencyCode"], 
					userFSM.UserData["secondCurrencyCode"],
					map_currencyes,
					amount,
				)
				if err != nil {
					msg.Text, event = "Ошибка конвертирования, попробуйте отправить число снова или отменити операцию /cancel", ""
				}else {
					msg.Text, event = userFSM.UserData["firstCurrencyCode"] + " на " + userFSM.UserData["secondCurrencyCode"] + " = " + answer, "finish"
				}
			}
		}

		userFSM.changeEvent(event)
		bot.Send(msg)
	}
}
