package converter

import (
	"mvasilyev/zamenabot/items"
	"mvasilyev/zamenabot/scheduler"
	"testing"
)

func TestConverter_Convert_Cancellation(t *testing.T) {
	converter := Converter{}

	// Test input with cancellation
	data := [][]string{
		{"01.02.2025", "1", "Mr. Smith", "John Doe", "Room 101", "", "отмена"},
	}

	expectedItems := []scheduler.ScheduleItem{
		&items.Cancelation{
			ScheduleItem: items.ScheduleItem{
				Date:        "01.02.2025",
				LessonNumber: 1,
				Teacher:     "Mr. Smith",
			},
		},
	}

	result := converter.Convert(data)

	// Compare lengths first
	if len(result) != len(expectedItems) {
		t.Fatalf("Expected %d items, got %d", len(expectedItems), len(result))
	}

	// Compare individual items
	for i, item := range result {
		if item.GetString() != expectedItems[i].GetString() {
			t.Errorf("Expected item %v, got %v", expectedItems[i], item)
		}
	}
}

func TestConverter_Convert_Substitution(t *testing.T) {
	converter := Converter{}

	// Test input with substitution
	data := [][]string{
		{"02.02.2025", "2", "Ms. Johnson", "Mr. Brown", "Room 202", "", "substitution"},
	}

	expectedItems := []scheduler.ScheduleItem{
		&items.Substitution{
			ScheduleItem: items.ScheduleItem{
				Date:        "02.02.2025",
				LessonNumber: 2,
				Teacher:     "Ms. Johnson",
			},
			SubstituteTeacher: "Mr. Brown",
			ClassRoom:         "Room 202",
		},
	}

	result := converter.Convert(data)

	// Compare lengths first
	if len(result) != len(expectedItems) {
		t.Fatalf("Expected %d items, got %d", len(expectedItems), len(result))
	}

	// Compare individual items
	for i, item := range result {
		if item.GetString() != expectedItems[i].GetString(){
			t.Errorf("Expected item %v, got %v", expectedItems[i], item)
		}
	}
}

func TestConverter_Convert_MixedData(t *testing.T) {
	converter := Converter{}

	// Test input with both cancellation and substitution
	data := [][]string{
		{"01.02.2025", "1", "Mr. Smith", "John Doe", "Room 101", "", "отмена"},
		{"02.02.2025", "2", "Ms. Johnson", "Mr. Brown", "Room 202", "", "substitution"},
	}

	expectedItems := []scheduler.ScheduleItem{
		&items.Cancelation{
			ScheduleItem: items.ScheduleItem{
				Date:        "01.02.2025",
				LessonNumber: 1,
				Teacher:     "Mr. Smith",
			},
		},
		&items.Substitution{
			ScheduleItem: items.ScheduleItem{
				Date:        "02.02.2025",
				LessonNumber: 2,
				Teacher:     "Ms. Johnson",
			},
			SubstituteTeacher: "Mr. Brown",
			ClassRoom:         "Room 202",
		},
	}

	result := converter.Convert(data)

	// Compare lengths first
	if len(result) != len(expectedItems) {
		t.Fatalf("Expected %d items, got %d", len(expectedItems), len(result))
	}

	// Compare individual items
	for i, item := range result {
		if item.GetString() != expectedItems[i].GetString() {
			t.Errorf("Expected item %v, got %v", expectedItems[i], item)
		}
	}
}

func TestConverter_Convert_EmptyData(t *testing.T) {
	converter := Converter{}

	// Test input with empty data
	data := [][]string{}

	expectedItems := []scheduler.ScheduleItem{}

	result := converter.Convert(data)

	// Compare lengths first
	if len(result) != len(expectedItems) {
		t.Fatalf("Expected %d items, got %d", len(expectedItems), len(result))
	}
}