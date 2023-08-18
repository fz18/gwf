package gwf

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

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

type LogFormatterParams struct {
	Request    *http.Request
	TimeStamp  time.Time
	StatusCode int
	Latency    time.Duration
	ClientIP   net.IP
	Method     string
	Path       string
}

type LoggerFormatter = func(params *LogFormatterParams) string

type LoggerConfig struct {
	Formatter LoggerFormatter
	Out       io.Writer
}

var defaultWriter io.Writer = os.Stdout

var defaultFormatter LoggerFormatter = func(params *LogFormatterParams) string {
	if params.Latency > time.Minute {
		params.Latency = params.Latency.Truncate(time.Second)
	}
	statusColor := statusColor(params.StatusCode)
	return fmt.Sprintf("%s [gwf] %s|%s %v %s|%s %3d %s|%s %13v %s| %15s |%s %-7s %s %s %#v %s\n", yellow, reset,
		blue, params.TimeStamp.Format("2006-01-02 15:04:05"), reset, statusColor,
		params.StatusCode, reset, red, params.Latency, reset,
		params.ClientIP, magenta, params.Method, reset, cyan, params.Path, reset)
}

func statusColor(status int) string {
	switch status {
	case http.StatusOK:
		return green
	default:
		return red
	}
}

func LoggerWithConfig(conf LoggerConfig, next HandlerFunc) HandlerFunc {
	formatter := conf.Formatter
	if formatter == nil {
		formatter = defaultFormatter
	}
	out := conf.Out
	if out == nil {
		out = defaultWriter
	}
	return func(c *Context) {

		start := time.Now()
		path := c.R.URL.Path
		raw := c.R.URL.RawQuery
		next(c)
		end := time.Now()
		latency := end.Sub(start)
		ip, _, _ := net.SplitHostPort(strings.TrimSpace(c.R.RemoteAddr))
		clientIP := net.ParseIP(ip)
		method := c.R.Method
		statusCode := c.StatusCode
		if raw != "" {
			path = path + "?" + raw
		}
		param := LogFormatterParams{
			Request:    c.R,
			StatusCode: statusCode,
			Latency:    latency,
			ClientIP:   clientIP,
			Method:     method,
			Path:       path,
			TimeStamp:  end,
		}
		fmt.Fprint(out, formatter(&param))

	}
}

func Logger(next HandlerFunc) HandlerFunc {
	return LoggerWithConfig(LoggerConfig{}, next)
}
