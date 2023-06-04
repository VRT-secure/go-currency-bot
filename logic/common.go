package logic

import (
	"context"
	"fmt"
	"log"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/looplab/fsm"
)

var MainMenuKeyboard = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton(MainMenuKeyboard_fiatCurrency),
		tgbotapi.NewKeyboardButton(MainMenuKeyboard_calculateFiatCyrrencyes),
	),
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton(MainMenuKeyboard_goldMakhachkala),
	),
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton(MainMenuKeyboard_metalCBR),
	),
)

type UserFSM struct {
	ChatID   int64
	FSM      *fsm.FSM
	UserData map[string]string
	TimeRequest time.Time
	IsBlocked bool
}

func NewUserFSM(chatID int64) *UserFSM {
	u := &UserFSM{
		ChatID:   chatID,
		UserData: make(map[string]string),
	}

	u.FSM = fsm.NewFSM(
		"init",
		fsm.Events{
			{Name: Start, Src: []string{"init",
				ChoiceOperation, ChoiceOneFiatCurrency,
				ChoiceFirstFiatCurrency, ChoiceSecondFiatCurrency,
				ChoiceFiatAmount}, Dst: ChoiceOperation},
			{Name: CourseFiatCurrency, Src: []string{ChoiceOperation}, Dst: ChoiceOneFiatCurrency},
			{Name: FirstFiatCyrrency, Src: []string{ChoiceOperation}, Dst: ChoiceFirstFiatCurrency},
			{Name: SecondFiatCyrrency, Src: []string{ChoiceFirstFiatCurrency}, Dst: ChoiceSecondFiatCurrency},
			{Name: FiatAmount, Src: []string{ChoiceSecondFiatCurrency}, Dst: ChoiceFiatAmount},
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

func (u *UserFSM) ChangeEvent(event string) {
	if event != "" {
		err := u.FSM.Event(context.Background(), event)
		if err != nil {
			log.Printf("FSM error: %v", err)
		}
	}
}

func (u *UserFSM) IsSpam() bool {
	// Получение текущего времени
	currentTime := time.Now()

	// Вычисление разницы между текущим временем и сохраненным временем запроса
	diff := currentTime.Sub(u.TimeRequest)

	// Проверка, превышает ли разница установленный интервал
	return diff < (CountSeconds * time.Second)
}


func MainMenu(text string, keyboard tgbotapi.ReplyKeyboardMarkup) (tgbotapi.ReplyKeyboardMarkup, string) {
	return keyboard, text
}
