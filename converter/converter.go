package converter

import (
	"mvasilyev/zamenabot/items"
	"strconv"
	"mvasilyev/zamenabot/scheduler"
)

type Converter struct{

}

type ScheduleItem struct {
	Date string
	Subject string
	Teacher string
	LessonNumber int
}

func (converter Converter) Convert(data [][]string) []scheduler.ScheduleItem {
	items := []scheduler.ScheduleItem{}
	for _, row := range data {
		if len(row) > 6 && row[6] == "отмена" {
			item := composeCancelationItem(row)
			items = append(items, item)
		} else {
			item := composeSubstituteMessage(row)
			items = append(items, item)
		}
	}

	return items
}

func composeCancelationItem(row []string) scheduler.ScheduleItem {
	return items.Cancelation{
		ScheduleItem: items.ScheduleItem{
			Date: row[0],
			LessonNumber: func() int {
				num, _ := strconv.Atoi(row[1])
				return num
			}(),
			Teacher: row[2],
		},
	}
}

func composeSubstituteMessage(row []string) scheduler.ScheduleItem {
	return items.Substitution{
		ScheduleItem: items.ScheduleItem{
			Date: row[0],
			LessonNumber: func() int {
				num, _ := strconv.Atoi(row[1])
				return num
			}(),
			Teacher: row[2],
		},
		SubstituteTeacher: row[3],
		ClassRoom: row[4],
	}
}