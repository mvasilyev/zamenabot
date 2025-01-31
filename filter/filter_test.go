package filter

import (
	"testing"
	"time"
)

func TestFilter_FilterSchedule_Success(t *testing.T) {
	filter := Filter{
		ClassID: "Math101",
	}

	// Get today's date and compute future dates
	now := time.Now().Truncate(24 * time.Hour)
	oneDayLater := now.Add(24 * time.Hour)
	twoDaysLater := now.Add(48 * time.Hour)
	threeDaysLater := now.Add(72 * time.Hour)

	data := [][]string{
		{now.Add(24 * time.Hour).Format("02.01.2006"), "", "", "", "Math101"},    // 1 day later
		{now.Add(48 * time.Hour).Format("02.01.2006"), "", "", "", "Math101"},    // 2 days later
		{now.Add(72 * time.Hour).Format("02.01.2006"), "", "", "", "Science101"}, // 3 days later, different class
		{"", "", "", "", ""}, // Empty row
		{now.Add(72 * time.Hour).Format("02.01.2006"), "", "", "", "Math101"}, // 3 days later, correct class
	}

	expectedFutureRows := [][]string{
		{oneDayLater.Format("02.01.2006"), "", "", "", "Math101"},
		{twoDaysLater.Format("02.01.2006"), "", "", "", "Math101"},
		{threeDaysLater.Format("02.01.2006"), "", "", "", "Math101"},
	}

	result := filter.FilterSchedule(data)

	if len(result) != len(expectedFutureRows) {
		t.Fatalf("Expected %v rows, got %v", len(expectedFutureRows), len(result))
	}

	for i, row := range result {
		if len(row) != len(expectedFutureRows[i]) {
			t.Fatalf("Expected row %v to have %v columns, got %v", i, len(expectedFutureRows[i]), len(row))
		}
		for j, cell := range row {
			if cell != expectedFutureRows[i][j] {
				t.Fatalf("Expected cell [%v][%v] to be %v, got %v", i, j, expectedFutureRows[i][j], cell)
			}
		}
	}
}

func TestFilter_FilterSchedule_SkipInvalidClass(t *testing.T) {
	filter := Filter{
		ClassID: "Math101",
	}

	// Get today's date and compute future dates
	now := time.Now().Truncate(24 * time.Hour)
	oneDayLater := now.Add(24 * time.Hour)
	twoDaysLater := now.Add(48 * time.Hour)

	data := [][]string{
		{oneDayLater.Format("02.01.2006"), "", "", "", "Math101"},             // 1 day later
		{twoDaysLater.Format("02.01.2006"), "", "", "", "Math102"},            // Wrong class
		{now.Add(72 * time.Hour).Format("02.01.2006"), "", "", "", "Math101"}, // 3 days later
	}

	expectedFutureRows := [][]string{
		{oneDayLater.Format("02.01.2006"), "", "", "", "Math101"},
		{now.Add(72 * time.Hour).Format("02.01.2006"), "", "", "", "Math101"},
	}

	result := filter.FilterSchedule(data)

	if len(result) != len(expectedFutureRows) {
		t.Fatalf("Expected %v rows, got %v", len(expectedFutureRows), len(result))
	}

	for i, row := range result {
		if len(row) != len(expectedFutureRows[i]) {
			t.Fatalf("Expected row %v to have %v columns, got %v", i, len(expectedFutureRows[i]), len(row))
		}
		for j, cell := range row {
			if cell != expectedFutureRows[i][j] {
				t.Fatalf("Expected cell [%v][%v] to be %v, got %v", i, j, expectedFutureRows[i][j], cell)
			}
		}
	}
}

func TestFilter_FilterSchedule_SkipInvalidDate(t *testing.T) {
	filter := Filter{
		ClassID: "Math101",
	}

	// Get today's date and compute future dates
	now := time.Now().Truncate(24 * time.Hour)
	oneDayLater := now.Add(24 * time.Hour)
	twoDaysLater := now.Add(48 * time.Hour)

	data := [][]string{
		{oneDayLater.Format("02.01.2006"), "", "", "", "Math101"},  // 1 day later
		{"invalid-date", "", "", "", "Math101"},                    // Invalid date
		{twoDaysLater.Format("02.01.2006"), "", "", "", "Math101"}, // 2 days later
	}

	expectedFutureRows := [][]string{
		{oneDayLater.Format("02.01.2006"), "", "", "", "Math101"},
		{twoDaysLater.Format("02.01.2006"), "", "", "", "Math101"},
	}

	result := filter.FilterSchedule(data)

	if len(result) != len(expectedFutureRows) {
		t.Fatalf("Expected %v rows, got %v", len(expectedFutureRows), len(result))
	}

	for i, row := range result {
		if len(row) != len(expectedFutureRows[i]) {
			t.Fatalf("Expected row %v to have %v columns, got %v", i, len(expectedFutureRows[i]), len(row))
		}
		for j, cell := range row {
			if cell != expectedFutureRows[i][j] {
				t.Fatalf("Expected cell [%v][%v] to be %v, got %v", i, j, expectedFutureRows[i][j], cell)
			}
		}
	}
}
