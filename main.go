package main

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	"log"
	"os"
	"strings"
	"time"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_APITOKEN"))
	if err != nil {
		panic(err)
	}

	DB := DB{}
	err = DB.init("test.db")
	if err != nil {
		panic(err)
	}

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 30
	updates := bot.GetUpdatesChan(updateConfig)

	go func() {
		for {
			time.Sleep(1 * time.Hour)
			ids, err := DB.deleteOldMessages(2)
			if err != nil {
				log.Println(err)
			}

			for _, bufArray := range ids {
				req := tgbotapi.NewDeleteMessage(bufArray[0], int(bufArray[1]))
				_, err := bot.Request(req)
				if err != nil {
					log.Println(err)
				}
			}
		}
	}()

	for update := range updates {
		if update.Message == nil {
			continue
		}

		if !update.Message.IsCommand() {
			continue
		}

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")

		switch update.Message.Command() {
		case "set":
			splitMsg := strings.Split(update.Message.Text, " ")

			if len(splitMsg) != 4 {
				msg.Text = "Неправильно введена команда"
			} else {
				messageObj := Message{
					UserId:      update.SentFrom().ID,
					ServiceName: splitMsg[1],
					Login:       splitMsg[2],
					Password:    splitMsg[3],
					AddDate:     time.Now().Unix(),
				}

				err := DB.addMessage(messageObj)
				if err != nil {
					if strings.Contains(err.Error(), "UNIQUE constraint failed") {
						msg.Text = "Такая запись уже присутсвует"
					} else {
						msg.Text = "Произашла ошибка на сервере. Поробуйте позже"
					}
					log.Println(err)
				} else {
					msg.Text = "Сервис успешно добавлен"
				}
			}
		case "get":
			splitMsg := strings.Split(update.Message.Text, " ")
			if len(splitMsg) != 2 {
				msg.Text = "Неправильно введена команда"
			} else {
				messageObj, err := DB.getMessage(int(update.SentFrom().ID), splitMsg[1])
				if err != nil {
					log.Println(err)
					msg.Text = "Произашла ошибка на сервере. Поробуйте позже"
				}

				if messageObj.Login == "" {
					msg.Text = "Запись не найдена"
				} else {
					msg.Text = messageObj.Login + " " + messageObj.Password
				}
			}
		case "del":
			splitMsg := strings.Split(update.Message.Text, " ")
			if len(splitMsg) != 2 {
				msg.Text = "Неправильно введена команда"
			} else {
				err := DB.deleteMessage(int(update.SentFrom().ID), splitMsg[1])
				if err != nil {
					log.Println(err)
					msg.Text = "Произашла ошибка на сервере. Поробуйте позже"
				} else {
					msg.Text = "Запись успешно удалена"
				}
			}
		default:
			msg.Text = "Неизвестная команда"
		}

		botMsg, err := bot.Send(msg)
		if err != nil {
			log.Println(err)
			msg.Text = "Произашла ошибка на сервере. Поробуйте позже"
		}

		err = DB.addMessageIdToDelete(botMsg.MessageID, update.SentFrom().ID)
		if err != nil {
			log.Println(err)
			msg.Text = "Произашла ошибка на сервере. Поробуйте позже"
		}

		err = DB.addMessageIdToDelete(update.Message.MessageID, update.SentFrom().ID)
		if err != nil {
			log.Println(err)
			msg.Text = "Произашла ошибка на сервере. Поробуйте позже"
		}
	}
}
