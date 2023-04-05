package metrics

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

var (
	SendCall  prometheus.Counter
	CheckCall prometheus.Counter
)

type MetricServer struct {
	Host      string
	Port      string
	Path      string
	MetricLog *logrus.Logger
}

func NewMetricServer() *MetricServer {
	return &MetricServer{}
}

func (m *MetricServer) Start() {
	m.MetricLog.WithFields(logrus.Fields{
		"metric":    "Hash compute metric server",
		"host_port": fmt.Sprintf("%s:%s", m.Host, m.Port),
		"path":      fmt.Sprintf("/%s", m.Path),
	}).Debug("Server is staring")

	SendCall = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "send_call",
		Help: "counter for calls POST /send",
	})
	CheckCall = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "check_call",
		Help: "counter for calls POST /check",
	})
	prometheus.Register(SendCall)
	prometheus.Register(CheckCall)

	http.Handle(fmt.Sprintf("/%s", m.Path), promhttp.Handler())

	server := &http.Server{Addr: fmt.Sprintf("%s:%s", m.Host, m.Port), Handler: nil}
	serverCtx, serverStopCtx := context.WithCancel(context.Background())

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-sig
		shutdownCtx, _ := context.WithTimeout(serverCtx, 1*time.Second)
		go func() {
			<-shutdownCtx.Done()
			if shutdownCtx.Err() == context.DeadlineExceeded {
				m.MetricLog.Fatal("graceful shutdown timed out.. forcing exit.")
			}
		}()
		err := server.Shutdown(shutdownCtx)
		if err != nil {
			m.MetricLog.Fatal(err)
		}
		m.MetricLog.Info("Metric server graceful shutdown")
		serverStopCtx()
	}()

	m.MetricLog.Infof("metric web server start on http://%s:%s/%s", m.Host, m.Port, m.Path)
	err := server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		m.MetricLog.Fatal(err)
	}

	<-serverCtx.Done()

}
