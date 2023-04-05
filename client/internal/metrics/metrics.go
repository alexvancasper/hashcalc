package metrics

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/go-chi/chi"
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
	// wg := new(sync.WaitGroup)
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
		shutdownCtx, _ := context.WithTimeout(serverCtx, 5*time.Second)
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

	// wg.Add(1)
	// go func() {
	// 	defer wg.Done()
	// 	err := http.ListenAndServe(fmt.Sprintf("%s:%s", m.Host, m.Port), nil)
	// 	if err != nil {
	// 		m.MetricLog.WithFields(logrus.Fields{
	// 			"metric":           "Hash compute metric server",
	// 			"host:port params": fmt.Sprintf("%s:%s", m.Host, m.Port),
	// 		}).Fatalf("metric server error %v", err)
	// 	}
	// }()
	// wg.Wait()
}

type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (rec *statusRecorder) WriteHeader(statusCode int) {
	rec.statusCode = statusCode
	rec.ResponseWriter.WriteHeader(statusCode)
}

func measureResponseDuration(next http.Handler) http.Handler {
	buckets := []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10}

	responseTimeHistogram := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "namespace",
		Name:      "http_server_request_duration_seconds",
		Help:      "Histogram of response time for handler in seconds",
		Buckets:   buckets,
	}, []string{"route", "method", "status_code"})

	prometheus.MustRegister(responseTimeHistogram)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := statusRecorder{w, 200}

		next.ServeHTTP(&rec, r)

		duration := time.Since(start)
		statusCode := strconv.Itoa(rec.statusCode)
		route := getRoutePattern(r)
		responseTimeHistogram.WithLabelValues(route, r.Method, statusCode).Observe(duration.Seconds())
	})
}

func getRoutePattern(r *http.Request) string {
	reqContext := chi.RouteContext(r.Context())
	if pattern := reqContext.RoutePattern(); pattern != "" {
		return pattern
	}

	return "undefined"
}
