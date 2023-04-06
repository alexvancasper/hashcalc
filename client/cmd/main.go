package main

import (
	"fmt"
	"hash-service-client/internal/config"
	"hash-service-client/internal/routes"
	"log"
	"os"

	metrics "hash-service-client/internal/metrics"

	formatter "github.com/fabienm/go-logrus-formatters"
	graylog "github.com/gemnasium/logrus-graylog-hook/v3"

	"github.com/sirupsen/logrus"
)

var cfg *config.Config
var MyLogger = logrus.New()

func init() {
	var err error
	fmt.Printf("%+v\n", os.Args)
	if len(os.Args) < 2 {
		fmt.Println("Usage: ./web <path to config>")
		os.Exit(1)
	}
	ConfigPath := os.Args[len(os.Args)-1]
	cfg, err = config.NewConfig(ConfigPath)
	if err != nil {
		log.Fatal(err)
	}

	gelfFmt := formatter.NewGelf("Compute web server")
	hook := graylog.NewGraylogHook(fmt.Sprintf("%s:%s", cfg.Logging.Host, cfg.Logging.Port), map[string]interface{}{})
	MyLogger.AddHook(hook)
	MyLogger.SetFormatter(gelfFmt)
	MyLogger.SetOutput(os.Stdout)
	MyLogger.SetLevel(logrus.InfoLevel)
}

func main() {
	ms := metrics.NewMetricServer()
	ms.Host = cfg.Metrics.Host
	ms.Port = cfg.Metrics.Port
	ms.Path = cfg.Metrics.Path
	ms.MetricLog = MyLogger
	go ms.Start()

	routes.Start(MyLogger, cfg)

}
