package gwf

import (
	"fmt"
	"net/http"
	"runtime"
	"strings"
)

func detail(err any) string {
	var pcs [32]uintptr
	n := runtime.Callers(4, pcs[:])
	var sb strings.Builder
	fmt.Fprintf(&sb, "%v\n", err)
	for _, pc := range pcs[0:n] {
		fn := runtime.FuncForPC(pc)
		file, line := fn.FileLine(pc)
		fmt.Fprintf(&sb, "\n\t%s:%d", file, line)
	}
	return sb.String()
}

func Recover(next HandlerFunc) HandlerFunc {
	return func(c *Context) {
		defer func() {
			if e := recover(); e != nil {
				c.Logger.Error(detail(e))
				c.Fail(http.StatusInternalServerError, "internal error")
			}
		}()
		next(c)
	}
}
