package scheduler

import (
	"fmt"
	"time"
)

const (
	InvalidRequestError  string = "InvalidRequest"
	ServiceInternalError string = "ServiceInternal"
	ValidationError      string = "Validation"
)

//GenerateOneTimeCronExpression a cron(..) expression for a single instance at Now+minutesFromNow
func GenerateOneTimeCronExpression(secFromNow int, t time.Time) string {
	a := t.Add(time.Second * time.Duration(secFromNow))
	return fmt.Sprintf("cron(%02d %02d %02d %02d ? %d)", a.Minute(), a.Hour(), a.Day(), a.Month(), a.Year())
}
