package items

import "fmt"

type Cancelation struct {
	ScheduleItem
}

func (cancelation Cancelation) GetString() string {
	return fmt.Sprintf("🚫 Отмена %s: %s\n%s(%d урок)", cancelation.Date, cancelation.Teacher, cancelation.Subject, cancelation.LessonNumber)
}