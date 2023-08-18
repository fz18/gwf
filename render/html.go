package render

import "net/http"

type HTML struct {
	Data string
}

func (h *HTML) Rend(w http.ResponseWriter, code int) error {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(code)
	_, err := w.Write([]byte(h.Data))
	return err
}
