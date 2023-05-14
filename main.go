package main

import (
	"log"
	"os"
	"strconv"

	"kakafoni/database"
	"kakafoni/fiat_currency"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	token := os.Getenv("TELEGRAM_APITOKEN")
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic("Ошибка, нет катой переменной окужения: ", err)
	}

	debug_bot := os.Getenv("DEBUG_BOT")
	bot.Debug, err = strconv.ParseBool(debug_bot)
	if err != nil {
		log.Panic("Ошибка, нет катой переменной окужения: ", err)
	}

	// создаём или открываем файл БД
	db, err := gorm.Open(sqlite.Open("currencyes.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	err = database.CreateTable(db, &fiat_currency.FiatCurrency{})
	if err != nil {
		panic("failed to migrate database")
	}

	if fiat_currency.IsFiatTableEmpty(db) {
		fiat_currency.ParseJsonIntoTable(db, fiat_currency.URL_TO_JSON_FIAT)
	}

	// запускаем cron задачу, которая выполняется каждую полночь
	c := cron.New()
	c.AddFunc("0 0 0 * * *", func() {
		fiat_currency.ParseJsonIntoTable(db, fiat_currency.URL_TO_JSON_FIAT)
		log.Printf("\nСработал крон\n")
	})
	c.Start()

	updateConfig := tgbotapi.NewUpdate(-1)
	updateConfig.Timeout = 60

	updates := bot.GetUpdatesChan(updateConfig)

	// машина состояний
	userFSMs := make(map[int64]*fiat_currency.UserFSM)

	charCodes := fiat_currency.FiatCharCodes(db)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		chatID := update.Message.Chat.ID

		userFSM, ok := userFSMs[chatID]
		if !ok {
			userFSM = fiat_currency.NewUserFSM(chatID)
			userFSMs[chatID] = userFSM
		}

		var event string

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
		switch update.Message.Command() {
		case "start":
			event = "start"
			msg.ReplyMarkup, msg.Text = fiat_currency.MainMenu("Выберите операцию", fiat_currency.MainMenuKeyboard)
		case "help":
			msg.Text = "Бот для конвертирования валют. Если хотите венруться в меню выбора операций отправьте /cancel"
			bot.Send(msg)
			continue
		case "cancel":
			event = "start"
			userFSM.ChangeEvent(event)
			msg.ReplyMarkup, msg.Text = fiat_currency.MainMenu("Отмена операции", fiat_currency.MainMenuKeyboard)
			bot.Send(msg)
			continue
		default:
			event = ""
		}

		user_text := update.Message.Text
		if userFSM.FSM.Current() == "choiceOperation" {

			if user_text == "Узнать курс валюты" {
				msg.ReplyMarkup = fiat_currency.CharCodesKeyboard(charCodes)
				event = "courseCurrency"
				msg.Text = "Выберите валюту"
			} else if user_text == "Калькулятор валют" {
				msg.ReplyMarkup = fiat_currency.CharCodesKeyboard(charCodes)
				event = "firstCyrrency"
				msg.Text = "Выберите первую валюту"
			} else {
				msg.Text = fiat_currency.INCORRECT_OPERATION
			}
		} else if userFSM.FSM.Current() == "choiceOneCurrency" {
			fiatCurrency, err := fiat_currency.SelectFiatFromTable(db, user_text)
			if err != nil {
				msg.Text, event = fiat_currency.ERROR_MESSAGE, ""
				bot.Send(msg)
				continue
			}
			msg.Text, event = fiat_currency.HandleFiatCurrencyChoice(charCodes, user_text, fiatCurrency, "start")
			bot.Send(msg)
			if event == "" {
				continue
			}

			msg.ReplyMarkup, msg.Text = fiat_currency.MainMenu("Выберите операцию", fiat_currency.MainMenuKeyboard)

		} else if userFSM.FSM.Current() == "choiceFirstCurrency" {
			fiatCurrency, err := fiat_currency.SelectFiatFromTable(db, user_text)
			if err != nil {
				msg.Text, event = fiat_currency.ERROR_MESSAGE, ""
				bot.Send(msg)
				continue
			}
			msg.Text, event = fiat_currency.HandleFiatCurrencyChoice(charCodes, user_text, fiatCurrency, "secondCyrrency")
			bot.Send(msg)
			if event == "" {
				continue
			}

			userFSM.UserData["firstCurrencyCode"] = user_text
			msg.Text = "Выберите вторую валюту"

		} else if userFSM.FSM.Current() == "choiceSecondCurrency" {
			fiatCurrency, err := fiat_currency.SelectFiatFromTable(db, user_text)
			if err != nil {
				msg.Text, event = fiat_currency.ERROR_MESSAGE, ""
				bot.Send(msg)
				continue
			}
			msg.Text, event = fiat_currency.HandleFiatCurrencyChoice(charCodes, user_text, fiatCurrency, "amount")
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
				msg.Text, event = fiat_currency.INCORRECT_NUMBER, ""
			} else {

				fist_fiatCurrency, err := fiat_currency.SelectFiatFromTable(db, userFSM.UserData["firstCurrencyCode"])
				if err != nil {
					msg.Text, event = fiat_currency.ERROR_MESSAGE, ""
					bot.Send(msg)
					continue
				}

				second_fiatCurrency, err := fiat_currency.SelectFiatFromTable(db, userFSM.UserData["secondCurrencyCode"])
				if err != nil {
					msg.Text, event = fiat_currency.ERROR_MESSAGE, ""
					bot.Send(msg)
					continue
				}

				answer, err := fiat_currency.ConvertFiatCurrency(
					fist_fiatCurrency,
					second_fiatCurrency,
					amount,
				)
				if err != nil {
					msg.Text, event = fiat_currency.ERROR_CONVERT, ""
				} else {
					msg.Text, event = strconv.Itoa(amount)+" "+userFSM.UserData["firstCurrencyCode"]+" = "+answer+" "+userFSM.UserData["secondCurrencyCode"], "start"
					bot.Send(msg)
					msg.ReplyMarkup, msg.Text = fiat_currency.MainMenu("Выберите операцию", fiat_currency.MainMenuKeyboard)
				}
			}
		}

		userFSM.ChangeEvent(event)
		bot.Send(msg)
	}
}
