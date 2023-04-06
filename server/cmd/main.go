package main

import (
	"context"
	"fmt"
	LRUCache "hashserver/internal/cache"
	"hashserver/internal/config"
	psql "hashserver/internal/database"
	"hashserver/internal/handlers"
	"hashserver/pkg/hashcalc"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	formatter "github.com/fabienm/go-logrus-formatters"
	graylog "github.com/gemnasium/logrus-graylog-hook/v3"
	"github.com/sirupsen/logrus"

	metrics "hashserver/internal/metrics"

	"google.golang.org/grpc"
)

var MyLogger = logrus.New()
var cfg *config.Config

func main() {

	var err error
	fmt.Printf("%+v\n", os.Args)
	if len(os.Args) < 2 {
		fmt.Println("Usage: ./server <path to config>")
		os.Exit(1)
	}
	ConfigPath := os.Args[len(os.Args)-1]
	cfg, err = config.NewConfig(ConfigPath)
	if err != nil {
		log.Fatal(err)
	}

	gelfFmt := formatter.NewGelf(cfg.Server.Name)
	hook := graylog.NewGraylogHook(fmt.Sprintf("%s:%s", cfg.Logging.Host, cfg.Logging.Port), map[string]interface{}{})
	MyLogger.AddHook(hook)
	MyLogger.SetFormatter(gelfFmt)
	MyLogger.SetOutput(os.Stdout)
	MyLogger.SetLevel(logrus.Level(cfg.Logging.Level))

	dbPool, err := psql.New(cfg.Server.DSN, cfg.Server.DB.PoolCount)
	if err != nil {
		MyLogger.Fatal(err)
	}

	err = psql.MigrationUP(dbPool)
	if err != nil {
		MyLogger.Fatal(err)
	}

	ms := metrics.NewMetricServer()
	ms.Host = cfg.Metrics.Host
	ms.Port = cfg.Metrics.Port
	ms.Path = cfg.Metrics.Path
	ms.MetricLog = MyLogger
	go ms.Start()

	s := grpc.NewServer()
	server := &handlers.Server{}
	server.DB = dbPool
	server.Logger = MyLogger
	server.Cache = LRUCache.NewLRUCache(cfg.Server.CacheCount)
	server.Workers = cfg.Server.HashWorkers
	hashcalc.RegisterHashCalcServer(s, server)

	serverCtx, serverStopCtx := context.WithCancel(context.Background())
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-sig
		shutdownCtx, _ := context.WithTimeout(serverCtx, 30*time.Second)
		go func() {
			<-shutdownCtx.Done()
			if shutdownCtx.Err() == context.DeadlineExceeded {
				MyLogger.Fatal("graceful shutdown timed out.. forcing exit.")
			}
		}()
		s.GracefulStop()
		MyLogger.Infof("%s gracefully shutdown", cfg.Server.Name)
		serverStopCtx()
	}()

	listenAddr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	lis, err := net.Listen("tcp", listenAddr)
	if err != nil {
		MyLogger.Fatal(err)
	}

	MyLogger.Printf("Starting server %s", listenAddr)
	if err := s.Serve(lis); err != nil {
		MyLogger.Fatal(err)
	}
	<-serverCtx.Done()
}
