package log

import (
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"strings"
	"time"

	"github.com/fz18/gwf/internal/gwfstrings"
)

type Level int

const (
	LevelInfo Level = iota
	LevelError
	LevelDebug
)

func (l Level) Color() string {
	switch l {
	case LevelInfo:
		return green
	case LevelError:
		return red
	case LevelDebug:
		return cyan
	default:
		return ""
	}
}

func (l Level) String() string {
	switch l {
	case LevelInfo:
		return "INFO"
	case LevelError:
		return "ERROR"
	case LevelDebug:
		return "DEBUG"
	default:
		return ""
	}
}

const (
	greenBg   = "\033[97;42m"
	whiteBg   = "\033[90;47m"
	yellowBg  = "\033[90;43m"
	redBg     = "\033[97;41m"
	magentaBg = "\033[97;45m"
	cyanBg    = "\033[97;46m"
	green     = "\033[32m"
	white     = "\033[37m"
	yellow    = "\033[33m"
	red       = "\033[31m"
	blue      = "\033[34m"
	magenta   = "\033[35m"
	cyan      = "\033[36m"
	reset     = "\033[0m"
)

type Fields map[string]any

type FormatParam struct {
	Level   Level
	IsColor bool
	Msg     any
	Fields  Fields
}
type Formatter interface {
	Format(param *FormatParam) string
}

type LoggerWriter struct {
	Level  Level
	Writer io.Writer
}

type Logger struct {
	Level     Level
	Outs      []*LoggerWriter
	Formatter Formatter
	Fields    Fields
	LogPath   string
	FileSize  int64
}

func Default() *Logger {
	return &Logger{
		Level:     LevelInfo,
		Outs:      []*LoggerWriter{{Level: -1, Writer: os.Stdout}},
		Formatter: &TextFormatter{},
	}
}

func (l *Logger) SetPath(logPath string) {
	l.LogPath = logPath
	l.Outs = append(l.Outs, &LoggerWriter{
		Level:  -1,
		Writer: FileWriter(path.Join(logPath, "all.log")),
	})
	l.Outs = append(l.Outs, &LoggerWriter{
		Level:  LevelInfo,
		Writer: FileWriter(path.Join(logPath, "info.log")),
	})
	l.Outs = append(l.Outs, &LoggerWriter{
		Level:  LevelDebug,
		Writer: FileWriter(path.Join(logPath, "debug.log")),
	})
	l.Outs = append(l.Outs, &LoggerWriter{
		Level:  LevelError,
		Writer: FileWriter(path.Join(logPath, "error.log")),
	})
}

func (l *Logger) Info(msg any) {
	l.print(LevelInfo, msg)
}
func (l *Logger) Error(msg any) {
	l.print(LevelError, msg)
}
func (l *Logger) Debug(msg any) {
	l.print(LevelDebug, msg)
}

func (l *Logger) WithFields(fields Fields) *Logger {
	return &Logger{
		Level:     l.Level,
		Formatter: l.Formatter,
		Outs:      l.Outs,
		Fields:    fields,
	}
}

func (l *Logger) print(level Level, msg any) {
	if l.Level > level {
		return
	}
	param := FormatParam{
		Level:  level,
		Fields: l.Fields,
		Msg:    msg,
	}
	for _, out := range l.Outs {
		if out.Writer == os.Stdout {
			param.IsColor = true
		}
		str := l.Formatter.Format(&param)
		if out.Level == -1 || out.Level == level {
			fmt.Fprint(out.Writer, str)
		}
		l.checkFileSize(out)
	}
}

func (l *Logger) checkFileSize(out *LoggerWriter) {
	// 获取文件描述符
	// 获取文件大小，以及文件名，判断大小，创建新文件
	logFile := out.Writer.(*os.File)
	if logFile == nil {
		return
	}
	stat, err := logFile.Stat()
	if err != nil {
		log.Print(err)
		return
	}
	size := stat.Size()
	if l.FileSize == 0 {
		l.FileSize = 100 << 20
	}
	if size < l.FileSize {
		return
	}
	_, fileName := path.Split(stat.Name())
	fileName = fileName[0:strings.Index(fileName, ".")]

	writer := FileWriter(path.Join(l.LogPath, gwfstrings.JoinStrings(
		fileName, ".", time.Now().Unix(), ".log")))
	out.Writer = writer
}

func FileWriter(name string) io.Writer {
	w, err := os.OpenFile(name, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0664)
	if err != nil {
		panic(err)
	}
	return w
}
