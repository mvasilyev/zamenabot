package main

import (
	"crypto/sha256"
	"encoding/csv"
	"fmt"
	"github.com/joho/godotenv"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

var (
	botToken   = getEnv("BOT_TOKEN", "")                    // Get Bot Token from environment or fallback
	chatID     = getEnv("CHAT_ID", "")                      // Get Chat ID from environment or fallback
	sheetID    = getEnv("SHEET_ID", "")                     // Get Google Sheet ID from environment or fallback
	classID    = getEnv("CLASS_ID", "")                     // Default class ID "7Ю", can be overridden
	checkTimes = getEnv("CHECK_TIMES", "06:00,12:00,18:00") // Default check times
	sentHashes = make(map[string]bool)
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

	timesToCheck := strings.Split(checkTimes, ",")

	for {
		now := time.Now()

		for _, timeStr := range timesToCheck {
			checkTime, err := time.Parse("15:04", timeStr)
			if err != nil {
				fmt.Println("Error parsing check time:", err)
				continue
			}

			if now.Hour() == checkTime.Hour() && now.Minute() == checkTime.Minute() {
				checkForUpdates()
				time.Sleep(60 * time.Second) // Wait a minute to avoid checking every second
			}
		}
	}
}

func checkForUpdates() {
	fmt.Println("Checking for updates...")

	// Fetch the data and send messages
	data, err := fetchSheetData(sheetID)
	if err != nil {
		fmt.Println("Error fetching sheet:", err)
		return
	}

	filteredData := filterFutureRowsForOurClass(data, classID)

	var messageParts []string

	for _, row := range filteredData {
		messageParts = append(messageParts, composeMessage(row))
	}

	fullMessage := strings.Join(messageParts, "\n\n")

	// Check for duplicate before sending
	messageHash := hashMessage(fullMessage)
	if _, exists := sentHashes[messageHash]; exists {
		fmt.Println("Duplicate message detected, not sending.")
		return
	}

	// Store the message hash and send it
	sentHashes[messageHash] = true

	err = sendTelegramMessage(fullMessage)
	if err != nil {
		fmt.Println("Error sending message:", err)
	} else {
		fmt.Println("Message sent:", fullMessage)
	}
}

// fetchSheetData fetches the Google Sheet data as CSV.
func fetchSheetData(sheetID string) ([][]string, error) {
	url := fmt.Sprintf("https://docs.google.com/spreadsheets/d/%s/gviz/tq?tqx=out:csv", sheetID)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	reader := csv.NewReader(resp.Body)
	data, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	return data, nil
}

// parseDate converts "DD.MM.YYYY" to a time.Time object.
func parseDate(dateStr string) (time.Time, error) {
	return time.Parse("02.01.2006", dateStr)
}

// filterFutureRows keeps track of the last valid date and applies it to rows without a date.
func filterFutureRowsForOurClass(data [][]string, classID string) [][]string {
	var futureRows [][]string
	var lastDate time.Time
	today := time.Now().Truncate(24 * time.Hour) // Get today's date without time

	for _, row := range data {
		if len(row) == 0 { // Skip empty rows
			continue
		}

		rawDate := strings.TrimSpace(row[0])

		// If the row has a new date, parse and update lastDate
		if rawDate != "" {
			parsedDate, err := parseDate(rawDate)
			if err != nil {
				fmt.Println("Skipping invalid date:", rawDate)
				continue
			}
			lastDate = parsedDate
		}

		rawClass := strings.TrimSpace(row[4])
		if !strings.Contains(rawClass, classID) {
			fmt.Println("Skipping invalid class:", rawClass)
			continue
		}

		// If we have a valid lastDate, check if it's in the future
		if !lastDate.IsZero() && lastDate.After(today) {
			if row[0] == "" {
				row[0] = lastDate.Format("02.01.2006")
			}

			futureRows = append(futureRows, row)
		}
	}

	return futureRows
}

func sendTelegramMessage(text string) error {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken)

	data := url.Values{}
	data.Set("chat_id", chatID)
	data.Set("text", text)
	data.Set("parse_mode", "Markdown") // Optional: Enables basic formatting

	_, err := http.PostForm(apiURL, data)
	return err
}

func composeMessage(row []string) string {
	if len(row) > 6 && row[6] == "отмена" {
		return composeCancelMessage(row)
	}

	return composeSubstituteMessage(row)
}

func composeSubstituteMessage(row []string) string {
	return fmt.Sprintf("🔄 Замена %s: %s\n%s(%s урок), заменяет %s в каб. %s", row[0], row[2], row[3], row[1], row[6], row[7])
}

func composeCancelMessage(row []string) string {
	return fmt.Sprintf("🚫 Отмена %s: %s\n%s(%s урок)", row[0], row[2], row[3], row[1])
}

// hashMessage generates a SHA256 hash for the message text.
func hashMessage(text string) string {
	hash := sha256.New()
	hash.Write([]byte(text))
	return fmt.Sprintf("%x", hash.Sum(nil))
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
