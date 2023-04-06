package routes

import (
	"context"
	"fmt"
	"hash-service-client/internal/config"
	"hash-service-client/internal/handlers"
	MW "hash-service-client/internal/middleware"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/sirupsen/logrus"
)

func Start(logger *logrus.Logger, cfg *config.Config) {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.WithValue("logger", logger))
	r.Use(middleware.WithValue("config", cfg))
	r.Use(MW.UUIDMiddleware)
	r.Use(middleware.Recoverer)
	httpClient := &http.Client{Timeout: time.Second * 5}
	h := &handlers.Handler{HttpClient: httpClient}

	r.Get("/check", h.Check)
	r.Get("/send", h.Web)
	r.Post("/send", h.Send)

	server := &http.Server{Addr: fmt.Sprintf("%s:%s", cfg.Client.Host, cfg.Client.Port), Handler: r}
	serverCtx, serverStopCtx := context.WithCancel(context.Background())

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-sig
		shutdownCtx, _ := context.WithTimeout(serverCtx, 5*time.Second)
		go func() {
			<-shutdownCtx.Done()
			if shutdownCtx.Err() == context.DeadlineExceeded {
				logger.Fatal("graceful shutdown timed out.. forcing exit.")
			}
		}()
		err := server.Shutdown(shutdownCtx)
		if err != nil {
			logger.Fatal(err)
		}
		logger.Info("WebUI server graceful shutdown")
		serverStopCtx()
	}()

	logger.Infof("Starting WebUI on port %s:%s", cfg.Client.Host, cfg.Client.Port)
	err := server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		logger.Fatal(err)
	}

	<-serverCtx.Done()
}
