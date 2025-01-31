package sender

import (
	"fmt"
	"mvasilyev/zamenabot/deduplicator"
	"mvasilyev/zamenabot/scheduler"
	"net/http"
	"net/url"
)

type Sender struct {
	BotToken     string
	ChatID       string
	Deduplicator deduplicator.Deduplicator
}

type Deduplicator interface {
	ShouldSend(string) bool
}

func (sender Sender) SendMessage(items []scheduler.ScheduleItem) {
	messageToSend := composeFullMessage(items)
	if sender.Deduplicator.ShouldSend(messageToSend) {
		err := sender.sendTelegramMessage(messageToSend)
		if err != nil {
			fmt.Println("Error sending message:", err)
		} else {
			fmt.Println("Message sent:", messageToSend)
		}
	}

}

func (sender *Sender) sendTelegramMessage(text string) error {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", sender.BotToken)

	data := url.Values{}
	data.Set("chat_id", sender.ChatID)
	data.Set("text", text)
	data.Set("parse_mode", "Markdown") // Optional: Enables basic formatting

	_, err := http.PostForm(apiURL, data)
	return err
}

func composeFullMessage(items []scheduler.ScheduleItem) string {
	message := ""

	for _, item := range items {
		message += item.GetString() + "\n\n"
	}

	return message
}
