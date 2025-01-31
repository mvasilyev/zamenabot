package items

import "fmt"

type Substitution struct {
	ScheduleItem

	SubstituteTeacher string
	ClassRoom string
}

func (substitution *Substitution) GetString() string {
	return fmt.Sprintf("🔄 Замена %s: %s\n%s(%d урок), заменяет %s в каб. %s", substitution.Date, substitution.Teacher, substitution.Subject, substitution.LessonNumber, substitution.SubstituteTeacher, substitution.ClassRoom)
}