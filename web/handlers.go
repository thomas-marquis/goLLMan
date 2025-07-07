package web

import (
	"github.com/thomas-marquis/goLLMan/web/components"
	"net/http"
)

func IndexHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		components.Index("Toto").Render(r.Context(), w)
	}
}
