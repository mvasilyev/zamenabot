package main

import (
	"fmt"
	"mvasilyev/zamenabot/converter"
	"mvasilyev/zamenabot/deduplicator"
	"mvasilyev/zamenabot/fetcher"
	"mvasilyev/zamenabot/filter"
	"mvasilyev/zamenabot/scheduler"
	"mvasilyev/zamenabot/sender"
	"mvasilyev/zamenabot/timechecker"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

var (
	botToken   = getEnv("BOT_TOKEN", "")                    // Get Bot Token from environment or fallback
	chatID     = getEnv("CHAT_ID", "")                      // Get Chat ID from environment or fallback
	sheetID    = getEnv("SHEET_ID", "")                     // Get Google Sheet ID from environment or fallback
	classID    = getEnv("CLASS_ID", "")                     // Default class ID "7Ю", can be overridden
	checkTimes = getEnv("CHECK_TIMES", "06:00,12:00,18:00") // Default check times
)

func main() {
	// Check if command-line arguments override environment variables
	overrideFromArgs()

	// Use environment variables or fallback to default values
	if sheetID == "" {
		fmt.Println("SHEET_ID is required")
		return
	}
	if botToken == "" {
		fmt.Println("BOT_TOKEN is required")
		return
	}
	if chatID == "" {
		fmt.Println("CHAT_ID is required")
		return
	}
	if checkTimes == "" {
		fmt.Println("CHECK_TIMES is required")
		return
	}

	checker := timechecker.TimeChecker{
		CheckTimes: checkTimes,
	}

	fetcher := fetcher.Fetcher{
		SheetID: sheetID,
		HTTPClient: &http.Client{},
	}

	filter := filter.Filter{
		ClassID: classID,
	}

	converter := converter.Converter{}

	sender := sender.Sender{
		BotToken:     botToken,
		ChatID:       chatID,
		Deduplicator: deduplicator.Deduplicator{},
	}

	scheduler := scheduler.Scheduler{
		Config: scheduler.Config{
			BotToken:   botToken,
			ChatID:     chatID,
			SheetID:    sheetID,
			ClassID:    classID,
			CheckTimes: checkTimes,
		},
		Checker:   &checker,
		Fetcher:   &fetcher,
		Filter:    &filter,
		Converter: &converter,
		Sender:    sender,
	}

	scheduler.Run()
}

// getEnv retrieves an environment variable or returns a fallback value
func getEnv(key string, fallback string) string {
	// Load environment variables from .env file
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Println("Error loading .env file:", err)
	}
	value, exists := os.LookupEnv(key)
	if !exists {
		return fallback
	}
	return value
}

// overrideFromArgs allows overriding of environment variables with command-line arguments.
func overrideFromArgs() {
	for i, arg := range os.Args {
		if arg == "--bot-token" && i+1 < len(os.Args) {
			botToken = os.Args[i+1]
		} else if arg == "--chat-id" && i+1 < len(os.Args) {
			chatID = os.Args[i+1]
		} else if arg == "--sheet-id" && i+1 < len(os.Args) {
			sheetID = os.Args[i+1]
		} else if arg == "--class-id" && i+1 < len(os.Args) {
			classID = os.Args[i+1]
		} else if arg == "--check-times" && i+1 < len(os.Args) {
			checkTimes = os.Args[i+1]
		}
	}
}
