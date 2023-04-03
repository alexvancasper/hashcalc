package main

import (
	"fmt"
	LRUCache "hashserver/internal/cache"
	"hashserver/internal/config"
	psql "hashserver/internal/database"
	"hashserver/internal/handlers"
	"hashserver/pkg/hashcalc"
	"net"
	"os"

	formatter "github.com/fabienm/go-logrus-formatters"
	// graylog "github.com/gemnasium/logrus-graylog-hook/v3"
	"github.com/sirupsen/logrus"

	"google.golang.org/grpc"
)

var MyLogger = logrus.New()

const Port string = "50051"

func init() {
	gelfFmt := formatter.NewGelf("Compute hash server")
	// hook := graylog.NewGraylogHook("localhost:12201", map[string]interface{}{})
	// MyLogger.AddHook(hook)
	MyLogger.SetFormatter(gelfFmt)
	MyLogger.SetOutput(os.Stdout)
	MyLogger.SetLevel(logrus.InfoLevel)
}

func main() {

	cfg, err := config.InitConfig("APP")
	if err != nil {
		MyLogger.Panic(err)
	}
	if cfg.Port == "" {
		cfg.Port = Port
	}
	db, err := psql.New(cfg.DSN)
	if err != nil {
		MyLogger.Panic(err)
	}

	// repo := repositories.New(db)
	// router := gin.Default()
	// handlerService := handlers.New(router, repo)
	// handlerService.Registry()
	// _ = router.Run(fmt.Sprintf(":%s", cfg.Port))

	listenAddr := fmt.Sprintf(":%s", cfg.Port)
	MyLogger.Printf("Starting server %s", listenAddr)
	lis, err := net.Listen("tcp", listenAddr)
	if err != nil {
		MyLogger.Panic(err)
	}
	s := grpc.NewServer()
	server := &handlers.Server{}
	server.DB = db
	server.Logger = MyLogger
	server.Cache = LRUCache.NewLRUCache(100)
	hashcalc.RegisterHashCalcServer(s, server)
	if err := s.Serve(lis); err != nil {
		MyLogger.Panic(err)
	}
}
