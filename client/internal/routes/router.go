package routes

import (
	"hash-service-client/internal/handlers"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi"
)

func Start(port string) {
	r := chi.NewRouter()
	httpClient := &http.Client{Timeout: time.Second * 5}
	h := &handlers.Handler{HttpClient: httpClient}

	r.Get("/check", h.Check)
	r.Get("/send", h.Web)
	r.Post("/send", h.Send)

	log.Fatal(http.ListenAndServe(":"+port, r))
}
