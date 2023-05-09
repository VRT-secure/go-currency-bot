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
			msg.Text = "Бот для конвертирования валют. Если хотите венруться в меню выбора операций отправьте /cancel"
			bot.Send(msg)
			continue
		case "cancel":
			event = "start"
			userFSM.changeEvent(event)
			msg.Text = "Операция отменена"
			msg.ReplyMarkup = choiceOperationKeyboard
			bot.Send(msg)
			continue
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
			if event == "" {
				continue
			}

			msg.ReplyMarkup = choiceOperationKeyboard
			msg.Text = "Выберите операцию"

		} else if userFSM.FSM.Current() == "choiceFirstCurrency" {
			charCode := update.Message.Text
			msg.Text, event = handleCurrencyChoice(charCodes, charCode, map_currencyes, "secondCyrrency")
			bot.Send(msg)
			if event == "" {
				continue
			}

			userFSM.UserData["firstCurrencyCode"] = charCode
			msg.Text = "Выберите вторую валюту"

		} else if userFSM.FSM.Current() == "choiceSecondCurrency" {
			charCode := update.Message.Text
			msg.Text, event = handleCurrencyChoice(charCodes, charCode, map_currencyes, "amount")
			bot.Send(msg)
			if event == "" {
				continue
			}

			userFSM.UserData["secondCurrencyCode"] = charCode
			msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
			msg.Text = "Введите количество целым числом"

		} else if userFSM.FSM.Current() == "choiceAmount" {
			amount, err := strconv.Atoi(update.Message.Text)
			if err != nil {
				msg.Text, event = "Ошибка, введите целое число или отмените операцию /cancel", ""
			}else {
				answer, err := convertCurrency(
					userFSM.UserData["firstCurrencyCode"], 
					userFSM.UserData["secondCurrencyCode"],
					map_currencyes,
					amount,
				)
				if err != nil {
					msg.Text, event = "Ошибка конвертирования, попробуйте отправить число снова или отмените операцию /cancel", ""
				}else {
					msg.Text, event =  strconv.Itoa(amount) + " " + userFSM.UserData["firstCurrencyCode"] + " = " + answer + " " + userFSM.UserData["secondCurrencyCode"], "start"
					bot.Send(msg)
					msg.ReplyMarkup = choiceOperationKeyboard
					msg.Text = "Выберите операцию"
				}
			}
		}

		userFSM.changeEvent(event)
		bot.Send(msg)
	}
}
