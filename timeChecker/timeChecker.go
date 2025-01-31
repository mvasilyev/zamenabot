package timechecker

import (
	"fmt"
	"strings"
	"time"
)

type TimeChecker struct{
	CheckTimes string
}

func (checker TimeChecker) ShouldCheck() bool {
	timesToCheck := strings.Split(checker.CheckTimes, ",")

	for _, timeToCheck := range timesToCheck {
		checkTime, err := time.Parse("15:04", timeToCheck)
		now := time.Now()
	
		if err != nil {
			fmt.Println("Error parsing check time:", err)
			continue
		}
	
		if now.Hour() == checkTime.Hour() && now.Minute() == checkTime.Minute() {
			return true
		}
	}

	return true
}

func (checker TimeChecker) Sleep() {
	time.Sleep(60 * time.Second)
}