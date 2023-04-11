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
	"github.com/pkg/errors"
	"google.golang.org/grpc"

	"github.com/sirupsen/logrus"
)

func Start(logger *logrus.Logger, cfg *config.Config) {
	l := logger.WithFields(logrus.Fields{
		"service":  "WEBUI",
		"module":   "client",
		"function": "Start",
	})

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(MW.UUIDMiddleware)
	r.Use(middleware.Recoverer)

	cwt, stopConn := context.WithCancel(context.Background())
	conn, err := grpc.DialContext(cwt, fmt.Sprintf("%s:%s", cfg.Grpc.Host, cfg.Grpc.Port), grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(time.Duration(5*time.Second)))
	if err != nil {
		l.WithField("error", fmt.Sprintf("%v", errors.WithStack(err))).Panic("not able to connect to grpc")
		return
	}
	logger.Info("successfully connected to gRPC")
	h := handlers.NewHandler()
	h.Logger = logger
	h.Server = conn
	defer conn.Close()

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
				l.Fatal("graceful shutdown timed out.. forcing exit.")
			}
		}()
		err := server.Shutdown(shutdownCtx)
		if err != nil {
			l.Fatal(err)
		}
		stopConn()
		l.Info("WebUI server graceful shutdown")
		serverStopCtx()

	}()

	list_addr := fmt.Sprintf("%s:%s", cfg.Client.Host, cfg.Client.Port)
	l.WithField("addr", list_addr).Infof("Starting WebUI on %s", list_addr)
	err = server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		l.Fatal(err)
	}

	<-serverCtx.Done()
	<-cwt.Done()
}
