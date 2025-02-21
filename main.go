package main

import (
	"crypto/sha256"
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config holds the environment variables
type Config struct {
	BotToken   string
	ChatID     string
	TopicID    string
	SheetID    string
	ClassID    string
	CheckTimes []string
}

// Global cache to prevent duplicate messages
var sentHashes = make(map[string]bool)

func main() {
	cfg := loadConfig()
	validateConfig(cfg)

	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()
		checkForUpdatesAt(cfg, now)
	}
}

// loadConfig reads configuration from environment variables or CLI arguments
func loadConfig() *Config {
	_ = godotenv.Load(".env") // Load .env file if available

	return &Config{
		BotToken:   getEnvOrArg("BOT_TOKEN", "--bot-token"),
		ChatID:     getEnvOrArg("CHAT_ID", "--chat-id"),
		TopicID:    getEnvOrArg("TOPIC_ID", "--topic-id"),
		SheetID:    getEnvOrArg("SHEET_ID", "--sheet-id"),
		ClassID:    getEnvOrArg("CLASS_ID", "--class-id"),
		CheckTimes: strings.Split(getEnvOrArg("CHECK_TIMES", "--check-times", "06:00,12:00,18:00"), ","),
	}
}

// validateConfig ensures all required settings are present
func validateConfig(cfg *Config) {
	if cfg.SheetID == "" || cfg.BotToken == "" || cfg.ChatID == "" || cfg.ClassID == ""{
		log.Fatal("Missing required configuration (SHEET_ID, BOT_TOKEN, CHAT_ID, CLASS_ID)")
	}
}

// checkForUpdatesAt triggers checks at scheduled times
func checkForUpdatesAt(cfg *Config, now time.Time) {

	for _, timeStr := range cfg.CheckTimes {
		checkTime, err := time.Parse("15:04", timeStr)
		if err != nil {
			log.Printf("Error parsing check time %s: %v\n", timeStr, err)
			continue
		}

		if now.Hour() == checkTime.Hour() && now.Minute() == checkTime.Minute() {
			checkForUpdates(cfg)
		}
	}
}

func checkForUpdates(cfg *Config) {
	log.Println("Checking for updates...")

	data, err := fetchSheetData(cfg.SheetID)
	if err != nil {
		log.Println("Error fetching sheet:", err)
		return
	}

	filteredData := filterFutureRowsForOurClass(data, cfg.ClassID)
	if len(filteredData) == 0 {
		return
	}

	var messages []string
	for _, row := range filteredData {
		messages = append(messages, composeMessage(row))
	}

	fullMessage := strings.Join(messages, "\n\n")
	messageHash := hashMessage(fullMessage)

	if sentHashes[messageHash] {
		log.Println("Duplicate message detected, not sending.")
		return
	}

	sentHashes[messageHash] = true

	if err := sendTelegramMessage(cfg, fullMessage); err != nil {
		log.Println("Error sending message:", err)
	} else {
		log.Println("Message sent:", fullMessage)
	}
}

// fetchSheetData retrieves Google Sheet data as CSV
func fetchSheetData(sheetID string) ([][]string, error) {
	resp, err := http.Get(fmt.Sprintf("https://docs.google.com/spreadsheets/d/%s/gviz/tq?tqx=out:csv", sheetID))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return csv.NewReader(resp.Body).ReadAll()
}

// filterFutureRowsForOurClass filters future rows for a given class
func filterFutureRowsForOurClass(data [][]string, classID string) [][]string {
	var result [][]string
	var lastDate time.Time
	yesterday := time.Now().Add(-24 * time.Hour).Truncate(24 * time.Hour)

	for _, row := range data {
		if len(row) == 0 {
			continue
		}

		rawDate := strings.TrimSpace(row[0])
		if rawDate != "" {
			parsedDate, err := parseDate(rawDate)
			if err != nil {
				log.Printf("Skipping invalid date: %s\n", rawDate)
				continue
			}
			lastDate = parsedDate
		}

		if !strings.Contains(strings.TrimSpace(row[4]), classID) {
			continue
		}

		if !lastDate.IsZero() && lastDate.After(yesterday) {
			row[0] = lastDate.Format("02.01.2006")
			
			result = append(result, row)
		}
	}

	return result
}

// parseDate converts "DD.MM.YYYY" to time.Time
func parseDate(rawDate string) (time.Time, error) {
	datePart := strings.Split(rawDate, " ")[0]

	return time.Parse("02.01.2006", datePart)
}

// sendTelegramMessage sends a message via Telegram bot
func sendTelegramMessage(cfg *Config, text string) error {
	data := url.Values{
		"chat_id":    {cfg.ChatID},
		"text":       {text},
		"parse_mode": {"Markdown"},
	}
	if cfg.TopicID != "" {
		data.Set("message_thread_id", cfg.TopicID)
	}

	_, err := http.PostForm(fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", cfg.BotToken), data)
	return err
}

// composeMessage formats the message text
func composeMessage(row []string) string {
	if len(row) > 6 && row[6] == "отмена" {
		return fmt.Sprintf("🚫 Отмена %s: %s\n%s (%s урок)", row[0], row[2], row[3], row[1])
	}
	return fmt.Sprintf("🔄 Замена %s: %s\n%s (%s урок), заменяет %s в каб. %s", row[0], row[2], row[3], row[1], row[6], row[7])
}

// hashMessage generates a SHA256 hash
func hashMessage(text string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(text)))
}

// getEnvOrArg retrieves a value from env or CLI args
func getEnvOrArg(envKey, argKey string, fallback ...string) string {
	if value, exists := os.LookupEnv(envKey); exists {
		return value
	}

	for i, arg := range os.Args {
		if arg == argKey && i+1 < len(os.Args) {
			return os.Args[i+1]
		}
	}

	if len(fallback) > 0 {
		return fallback[0]
	}
	return ""
}
