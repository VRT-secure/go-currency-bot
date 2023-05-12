package main

import (
	"log"
	"os"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/robfig/cron/v3"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	token := os.Getenv("TELEGRAM_APITOKEN")
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	// создаём или открываем файл БД
	db, err := gorm.Open(sqlite.Open("currencyes.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	createTable(db, &FiatCurrency{})

	if isFiatTableEmpty(db) {
		parseJsonIntoTable(db, URL)
	}
	// запускаем cron задачу, которая выполняется каждую полночь
	c := cron.New()
	c.AddFunc("0 0 0 * * *", func() {
		parseJsonIntoTable(db, URL)
	})

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60

	updates := bot.GetUpdatesChan(updateConfig)

	// машина состояний
	userFSMs := make(map[int64]*UserFSM)

	charCodes := fiatCharCodes(db)

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
			msg.ReplyMarkup, msg.Text = mainMenu("Выберите операцию", mainMenuKeyboard)
		case "help":
			msg.Text = "Бот для конвертирования валют. Если хотите венруться в меню выбора операций отправьте /cancel"
			bot.Send(msg)
			continue
		case "cancel":
			event = "start"
			userFSM.changeEvent(event)
			msg.ReplyMarkup, msg.Text = mainMenu("Отмена операции", mainMenuKeyboard)
			bot.Send(msg)
			continue
		default:
			event = ""
		}

		user_text := update.Message.Text
		if userFSM.FSM.Current() == "choiceOperation" {

			if user_text == "Узнать курс валюты" {
				msg.ReplyMarkup = charCodesKeyboard(charCodes)
				event = "courseCurrency"
				msg.Text = "Выберите валюту"
			} else if user_text == "Калькулятор валют" {
				msg.ReplyMarkup = charCodesKeyboard(charCodes)
				event = "firstCyrrency"
				msg.Text = "Выберите первую валюту"
			} else {
				msg.Text = "Ошибка, такой операции не существует, попробуйте снова"
			}
		} else if userFSM.FSM.Current() == "choiceOneCurrency" {
			msg.Text, event = handleFiatCurrencyChoice(charCodes, user_text, map_currencyes, "start")
			bot.Send(msg)
			if event == "" {
				continue
			}

			msg.ReplyMarkup, msg.Text = mainMenu("Выберите операцию", mainMenuKeyboard)

		} else if userFSM.FSM.Current() == "choiceFirstCurrency" {
			msg.Text, event = handleFiatCurrencyChoice(charCodes, user_text, map_currencyes, "secondCyrrency")
			bot.Send(msg)
			if event == "" {
				continue
			}

			userFSM.UserData["firstCurrencyCode"] = user_text
			msg.Text = "Выберите вторую валюту"

		} else if userFSM.FSM.Current() == "choiceSecondCurrency" {
			msg.Text, event = handleFiatCurrencyChoice(charCodes, user_text, map_currencyes, "amount")
			bot.Send(msg)
			if event == "" {
				continue
			}

			userFSM.UserData["secondCurrencyCode"] = user_text
			msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
			msg.Text = "Введите количество целым числом"

		} else if userFSM.FSM.Current() == "choiceAmount" {
			amount, err := strconv.Atoi(user_text)
			if err != nil {
				msg.Text, event = "Ошибка, введите целое число или отмените операцию /cancel", ""
			} else {
				answer, err := convertFiatCurrency(
					userFSM.UserData["firstCurrencyCode"],
					userFSM.UserData["secondCurrencyCode"],
					map_currencyes,
					amount,
				)
				if err != nil {
					msg.Text, event = "Ошибка конвертирования, попробуйте отправить число снова или отмените операцию /cancel", ""
				} else {
					msg.Text, event = strconv.Itoa(amount)+" "+userFSM.UserData["firstCurrencyCode"]+" = "+answer+" "+userFSM.UserData["secondCurrencyCode"], "start"
					bot.Send(msg)
					msg.ReplyMarkup, msg.Text = mainMenu("Выберите операцию", mainMenuKeyboard)
				}
			}
		}

		userFSM.changeEvent(event)
		bot.Send(msg)
	}
}
