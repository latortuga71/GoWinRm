package log

import "fmt"

type Logger struct{}

var Log *Logger
var Debug bool = false

func init() {
	Log = NewLogger()
}

func NewLogger() *Logger {
	return &Logger{}
}
func (l *Logger) Log(message string) {
	if Debug == true {
		fmt.Printf("DEBUG: %s\n", message)
	}
}
