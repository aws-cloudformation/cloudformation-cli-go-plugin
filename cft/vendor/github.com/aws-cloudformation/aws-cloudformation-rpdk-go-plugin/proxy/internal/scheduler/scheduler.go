package scheduler

import (
	"crypto/rand"
	"fmt"
	"io"
	"time"
)

//GenerateOneTimeCronExpression a cron(..) expression for a single instance at Now+minutesFromNow
func GenerateOneTimeCronExpression(minutesFromNow int, t time.Time) string {
	a := t.Add(time.Minute * time.Duration(minutesFromNow))
	return fmt.Sprintf("cron(%02d %02d %02d %02d ? %d)", a.Minute(), a.Hour(), a.Day(), a.Month(), a.Year())
}

// NewUUID generates a random UUID according to RFC 4122.
func NewUUID() (string, error) {
	uuid := make([]byte, 16)
	n, err := io.ReadFull(rand.Reader, uuid)
	if n != len(uuid) || err != nil {
		return "", err
	}
	uuid[8] = uuid[8]&^0xc0 | 0x80
	uuid[6] = uuid[6]&^0xf0 | 0x40
	return fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:]), nil
}
