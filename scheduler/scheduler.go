package scheduler

type Config struct {
	BotToken   string
	ChatID     string
	SheetID    string
	ClassID    string
	CheckTimes string
}

type Scheduler struct {
	Config Config
	Checker TimeChecker
	Filter Filter
	Fetcher Fetcher
	Converter Converter
	Sender Sender
}

type ScheduleItem interface {
	GetString() string
}

type TimeChecker interface{
	ShouldCheck() bool
	Sleep()
}

type Fetcher interface{
	FetchSchedule() ([][]string, error)
}

type Filter interface{
	FilterSchedule([][]string) [][]string
}

type Converter interface{
	Convert([][]string) []ScheduleItem
}

type Sender interface{
	SendMessage([]ScheduleItem)
}

func (scheduler Scheduler) Run () {
	for {
		if (scheduler.Checker.ShouldCheck()) {
			schedule, err := scheduler.Fetcher.FetchSchedule()
			if (err != nil) {
				// log
				continue
			}

			filteredSchedule := scheduler.Filter.FilterSchedule(schedule)
			convertedSchedule := scheduler.Converter.Convert(filteredSchedule)

			if (filteredSchedule != nil) {
				scheduler.Sender.SendMessage(convertedSchedule)
			}
		} 
		
		scheduler.Checker.Sleep()
	}
}
