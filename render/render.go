package render

import "net/http"

type Render interface {
	Rend(http.ResponseWriter, int) error
}
