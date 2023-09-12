package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type bnResp struct {
	Price float64 `json:"price,string"`
	Code  int64   `json:"code"`
}
type wallet map[string]float64

var db = map[int64]wallet{}

func main() {
	bot, err := tgbotapi.NewBotAPI("MyAwesomeBotToken")
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil { // If we got a message
			continue
		}

		command := strings.Split(update.Message.Text, " ")

		switch command[0] {
		case "ADD":
			if len(command) != 3 {
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Неверная команда"))
			}
			amount, err := strconv.ParseFloat(command[2], 64)
			if err != nil {
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, err.Error()))
			}

			if _, ok := db[update.Message.Chat.ID]; !ok {
				db[update.Message.Chat.ID] = wallet{}
			}

			db[update.Message.Chat.ID][command[1]] += amount
			balanceText := fmt.Sprintf("%f", db[update.Message.Chat.ID][command[1]])
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, balanceText))
		case "SUB":

			if len(command) != 3 {
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Неверная команда"))
			}
			amount, err := strconv.ParseFloat(command[2], 64)
			if err != nil {
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, err.Error()))
			}

			if _, ok := db[update.Message.Chat.ID]; !ok {
				continue
			}

			db[update.Message.Chat.ID][command[1]] -= amount
			balanceText := fmt.Sprintf("%f", db[update.Message.Chat.ID][command[1]])
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, balanceText))
		case "DEL":
			if len(command) != 2 {
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Неверная команда"))
			}

			delete(db[update.Message.Chat.ID], command[1])
		case "SHOW":
			mgs := " "
			var sum float64 = 0
			for key, value := range db[update.Message.Chat.ID] {
				price, _ := getPrice(key)
				sum += value * price
				mgs += fmt.Sprintf("%s: %f [%f]\n", key, value, value*price)
			}
			mgs += fmt.Sprintf("Total:%.2f\n", sum)
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, mgs))

		default:
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Команда не найдена"))
		}

	}
}

func getPrice(symbol string) (price float64, err error) {
	resp, err := http.Get(fmt.Sprintf("https://api.binance.com/api/v3/ticker/price?symbol=%sUSDT", symbol))
	if err != nil {
		return
	}
	defer resp.Body.Close()

	var jsonResp bnResp

	err = json.NewDecoder(resp.Body).Decode(&jsonResp)
	if err != nil {
		return
	}
	if jsonResp.Code != 0 {
		err = errors.New("Неверный символ")
	}

	price = jsonResp.Price
	return
}
