package filter

import (
	"fmt"
	"strings"
	"time"
)

type Filter struct {
	ClassID string
}

func (filter Filter) FilterSchedule(data [][]string) [][]string {
	return filterFutureRowsForOurClass(data, filter.ClassID)
}


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
			parsedDate, err := time.Parse("02.01.2006", rawDate)
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