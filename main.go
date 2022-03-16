package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func getKey() string {
	return "5213292130:AAHPqx69mCiBMT5LoV1b9T-PU3ve1DNH3ak"
}

type wallet map[string]float64

type binanceResp struct {
	Price float64 `json:"price,string"`
	Code  int64   `json:"code"`
}

var db = map[int64]wallet{}

func main() {
	bot, err := tgbotapi.NewBotAPI(getKey())
	if err != nil {
		log.Panic(err)
	}

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, _ := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil { // If we got a message
			log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

			msgArr := strings.Split(update.Message.Text, " ")

			switch msgArr[0] {
			case "ADD":
				// && reflect.ValueOf(msgArr[1]).IsNil() && !reflect.ValueOf(msgArr[2]).IsNil()
				if len(msgArr) >= 3 {
					sum, err := strconv.ParseFloat(msgArr[2], 64)
					if err != nil {
						bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Не верный формат значения суммы: "+msgArr[2]))
						continue
					}
					if _, ok := db[update.Message.Chat.ID]; !ok {
						db[update.Message.Chat.ID] = wallet{}
					}

					db[update.Message.Chat.ID][msgArr[1]] += sum
					bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Валюта "+msgArr[1]+" добавлена"))
					bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Итого: "+fmt.Sprintf("%f", db[update.Message.Chat.ID][msgArr[1]])+" "+msgArr[1]))
				} else {
					bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Неверная команда"))
				}

			case "SUB":
				if len(msgArr) >= 3 {
					sum, err := strconv.ParseFloat(msgArr[2], 64)
					if err != nil {
						bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Не верный формат значения суммы: "+msgArr[2]))
						continue
					}
					if _, ok := db[update.Message.Chat.ID]; !ok {
						db[update.Message.Chat.ID] = wallet{}
					}

					db[update.Message.Chat.ID][msgArr[1]] -= sum
					bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Валюта "+msgArr[1]+" убавлена"))
					bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Итого: "+fmt.Sprintf("%f", db[update.Message.Chat.ID][msgArr[1]])+" "+msgArr[1]))
				} else {
					bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Неверная команда"))
				}
			case "DEL":
				delete(db[update.Message.Chat.ID], msgArr[1])
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Валюта "+msgArr[1]+" удалена"))

			case "SHOW":
				msg := "Ваш баланс: \n"
				var usdSumm float64
				for key, val := range db[update.Message.Chat.ID] {
					coinPrice, err := getPrice(key)
					usdSumm += val * coinPrice
					if err != nil {
						bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, err.Error()))
					}
					msg += fmt.Sprintf("%s: %f [$%.2f]\n", key, val, val*coinPrice)
				}
				msg += fmt.Sprintf("\nИТОГО: $%.2f\n", usdSumm)
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, msg))

			default:
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Неизвестная команда"))

			}

		}
	}
}

func getPrice(coin string) (price float64, err error) {
	resp, err := http.Get(fmt.Sprintf("https://api.binance.com/api/v3/ticker/price?symbol=%sUSDT", coin))
	if err != nil {
		return
	}
	defer resp.Body.Close()
	var jsonResp binanceResp
	err = json.NewDecoder(resp.Body).Decode(&jsonResp)
	if err != nil {
		return
	}
	if jsonResp.Code != 0 {
		err = errors.New("Некорректная валюта")
		return
	}
	price = jsonResp.Price
	return
}
