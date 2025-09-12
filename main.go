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

// Installation represents a single bot installation configuration
type Installation struct {
	ID            string
	BotToken      string
	ChatID        string
	TopicID       string
	SheetID       string
	ClassID       string
	CheckTimes    []string
	MorningCutoff string
	sentHashes    map[string]bool // Per-installation cache
}

// Config holds multiple installations
type Config struct {
	Installations []*Installation
}

func main() {
	cfg := loadConfig()
	
	if len(cfg.Installations) == 0 {
		log.Fatal("No installations configured")
	}

	// Start a goroutine for each installation
	for _, installation := range cfg.Installations {
		go runInstallation(installation)
	}

	// Keep the main goroutine alive
	select {}
}

// runInstallation runs a single installation in its own goroutine
func runInstallation(installation *Installation) {
	log.Printf("[%s] Starting installation for class %s", installation.ID, installation.ClassID)
	
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()
		checkForUpdatesAt(installation, now)
	}
}

// loadConfig reads configuration from environment variables or multiple .env files
func loadConfig() *Config {
	installations := []*Installation{}
	
	// Check for multiple .env files pattern
	envFiles := findEnvFiles()
	
	if len(envFiles) == 0 {
		// Fallback to single installation from default .env or environment
		installation := loadSingleInstallation(".env", "default")
		if installation != nil {
			installations = append(installations, installation)
		}
	} else {
		// Load multiple installations from .env files
		for _, envFile := range envFiles {
			installationID := extractInstallationID(envFile)
			installation := loadSingleInstallation(envFile, installationID)
			if installation != nil {
				installations = append(installations, installation)
			}
		}
	}

	return &Config{Installations: installations}
}

// findEnvFiles searches for .env.* files in the current directory
func findEnvFiles() []string {
	var envFiles []string
	
	entries, err := os.ReadDir(".")
	if err != nil {
		return envFiles
	}
	
	for _, entry := range entries {
		name := entry.Name()
		// Skip example files and directories
		if strings.HasPrefix(name, ".env.") && !strings.HasSuffix(name, ".example") && !entry.IsDir() {
			envFiles = append(envFiles, name)
		}
	}
	
	return envFiles
}

// extractInstallationID extracts installation ID from filename like .env.school1 -> school1
func extractInstallationID(filename string) string {
	parts := strings.Split(filename, ".")
	if len(parts) >= 3 {
		return parts[2] // .env.school1 -> school1
	}
	return filename
}

// loadSingleInstallation loads configuration for a single installation
func loadSingleInstallation(envFile, installationID string) *Installation {
	// Clear environment variables that might interfere
	envVarsToLoad := []string{"BOT_TOKEN", "CHAT_ID", "TOPIC_ID", "SHEET_ID", "CLASS_ID", "MORNING_CUTOFF", "CHECK_TIMES"}
	
	// Store current environment values
	oldEnvValues := make(map[string]string)
	for _, envVar := range envVarsToLoad {
		if value, exists := os.LookupEnv(envVar); exists {
			oldEnvValues[envVar] = value
		}
		os.Unsetenv(envVar) // Clear the variable
	}
	
	// Load the specific .env file
	_ = godotenv.Load(envFile)

	installation := &Installation{
		ID:            installationID,
		BotToken:      getEnvOrArg("BOT_TOKEN", "--bot-token"),
		ChatID:        getEnvOrArg("CHAT_ID", "--chat-id"),
		TopicID:       getEnvOrArg("TOPIC_ID", "--topic-id"),
		SheetID:       getEnvOrArg("SHEET_ID", "--sheet-id"),
		ClassID:       getEnvOrArg("CLASS_ID", "--class-id"),
		MorningCutoff: getEnvOrArg("MORNING_CUTOFF", "--morning-cutoff", "12:00"),
		CheckTimes:    strings.Split(getEnvOrArg("CHECK_TIMES", "--check-times", "06:00,12:00,18:00"), ","),
		sentHashes:    make(map[string]bool),
	}
	
	// Restore old environment values
	for _, envVar := range envVarsToLoad {
		os.Unsetenv(envVar) // Clear again
		if value, exists := oldEnvValues[envVar]; exists {
			os.Setenv(envVar, value)
		}
	}
	
	// Validate required fields
	if installation.SheetID == "" || installation.BotToken == "" || installation.ChatID == "" || installation.ClassID == "" {
		if envFile == ".env" {
			log.Fatal("Missing required configuration (SHEET_ID, BOT_TOKEN, CHAT_ID, CLASS_ID)")
		} else {
			log.Printf("Skipping %s: missing required configuration", envFile)
			return nil
		}
	}
	
	log.Printf("Loaded installation %s for class %s", installation.ID, installation.ClassID)
	return installation
}

// checkForUpdatesAt triggers checks at scheduled times
func checkForUpdatesAt(installation *Installation, now time.Time) {
	for _, timeStr := range installation.CheckTimes {
		checkTime, err := time.Parse("15:04", timeStr)
		if err != nil {
			log.Printf("[%s] Error parsing check time %s: %v", installation.ID, timeStr, err)
			continue
		}

		if now.Hour() == checkTime.Hour() && now.Minute() == checkTime.Minute() {
			checkForUpdates(installation)
		}
	}
}

func checkForUpdates(installation *Installation) {
	log.Printf("[%s] Checking for updates...", installation.ID)

	data, err := fetchSheetData(installation.SheetID)
	if err != nil {
		log.Printf("[%s] Error fetching sheet: %v", installation.ID, err)
		return
	}

	filteredData := filterFutureRowsForOurClass(data, installation.ClassID, installation.MorningCutoff)
	if len(filteredData) == 0 {
		return
	}

	var messages []string
	for _, row := range filteredData {
		messages = append(messages, composeMessage(row))
	}

	fullMessage := strings.Join(messages, "\n\n")
	messageHash := hashMessage(fullMessage)

	if installation.sentHashes[messageHash] {
		log.Printf("[%s] Duplicate message detected, not sending.", installation.ID)
		return
	}

	installation.sentHashes[messageHash] = true

	if err := sendTelegramMessage(installation, fullMessage); err != nil {
		log.Printf("[%s] Error sending message: %v", installation.ID, err)
	} else {
		log.Printf("[%s] Message sent: %s", installation.ID, fullMessage)
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

func isMorningUpdate(morningCutoff string) bool {
	now := time.Now()

	cutoffTime, err := time.Parse("15:04", morningCutoff)
	if err != nil {
		log.Printf("Invalid MORNING_CUTOFF format: %s\n", morningCutoff)
		return false
	}

	morningDeadline := time.Date(now.Year(), now.Month(), now.Day(), cutoffTime.Hour(), cutoffTime.Minute(), 0, 0, now.Location())

	return now.Before(morningDeadline)
}

// filterFutureRowsForOurClass filters future rows for a given class
func filterFutureRowsForOurClass(data [][]string, classID string, morningCutoff string) [][]string {
	var result [][]string
	var lastDate time.Time

	now := time.Now()
	isMorning := isMorningUpdate(morningCutoff)

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

		if !lastDate.IsZero() {
			if isMorning && !lastDate.Before(now.Truncate(24*time.Hour)) {
				row[0] = lastDate.Format("02.01.2006")
				result = append(result, row)
			}

			if !isMorning && !lastDate.Before(now.Add(24*time.Hour).Truncate(24*time.Hour)) {
				row[0] = lastDate.Format("02.01.2006")
				result = append(result, row)
			}
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
func sendTelegramMessage(installation *Installation, text string) error {
	data := url.Values{
		"chat_id":    {installation.ChatID},
		"text":       {text},
		"parse_mode": {"Markdown"},
	}
	if installation.TopicID != "" {
		data.Set("message_thread_id", installation.TopicID)
	}

	_, err := http.PostForm(fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", installation.BotToken), data)
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
