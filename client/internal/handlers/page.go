package handlers

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type Handler struct {
	Logger *logrus.Logger
	Server *grpc.ClientConn
}

func (h *Handler) Web(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("web/templates/send.html")
	if err != nil {
		fmt.Println(err)
	}

	t.Execute(w, nil)
}

func NewHandler() *Handler {
	return &Handler{
		Logger: nil,
		Server: nil,
	}
}
