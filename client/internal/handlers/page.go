package handlers

import (
	"fmt"
	"html/template"
	"net/http"
)

type Handler struct {
	HttpClient *http.Client
}

func (h *Handler) Web(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("web/templates/send.html")
	if err != nil {
		fmt.Println(err)
	}

	t.Execute(w, nil)
}
