package display

import (
	"fmt"
	"time"
)

const TimeFormat = "15:04:05.000"

func PrintfWithTime(format string, args ...any) {
	args = append([]interface{}{time.Now().Format(TimeFormat)}, args...)
	fmt.Printf("%s "+format, args...)
}

func PrintlnWithTime(args ...any) {
	args = append([]interface{}{time.Now().Format(TimeFormat)}, args...)
	fmt.Println(args...)
}
