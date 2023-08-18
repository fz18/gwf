package log

import (
	"fmt"
	"time"
)

type TextFormatter struct{}

func (f *TextFormatter) Format(param *FormatParam) string {
	if param.IsColor {
		return fmt.Sprintf("%s [gwf] %s | %s | %s %s %s | %s %v %s %v\n", yellow, reset,
			time.Now().Format("2006-01-02 15:04:05"), param.Level.Color(), param.Level, reset,
			param.Level.Color(), param.Msg, reset, param.Fields)
	}
	return fmt.Sprintf("[gwf] | %s | %s | %v %v\n",
		time.Now().Format("2006-01-02 15:04:05"), param.Level, param.Msg, param.Fields)
}
