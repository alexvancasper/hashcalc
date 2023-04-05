package main

import (
	"hash-service-client/internal/routes"
	"os"
	"syscall"

	formatter "github.com/fabienm/go-logrus-formatters"

	// graylog "github.com/gemnasium/logrus-graylog-hook/v3"
	metrics "hash-service-client/internal/metrics"

	"github.com/sirupsen/logrus"
)

const DefaultPort = "8070"

var MyLogger = logrus.New()

func init() {
	gelfFmt := formatter.NewGelf("Compute web server")
	// hook := graylog.NewGraylogHook("localhost:12201", map[string]interface{}{})
	// MyLogger.AddHook(hook)
	MyLogger.SetFormatter(gelfFmt)
	MyLogger.SetOutput(os.Stdout)
	MyLogger.SetLevel(logrus.InfoLevel)
}

func main() {
	ms := metrics.NewMetricServer()
	ms.Host = ""
	ms.Port = "7766"
	ms.Path = "metrics"
	ms.MetricLog = MyLogger
	go ms.Start()

	port := DefaultPort
	if value, ok := syscall.Getenv("CS_PORT"); ok {
		port = value
	}

	routes.Start(MyLogger, port)

}
