package main

import (
	"context"
	"database/sql"
	"fmt"
	LRUCache "hashserver/internal/cache"
	"hashserver/internal/config"
	psql "hashserver/internal/database"
	"hashserver/internal/handlers"
	metrics "hashserver/internal/metrics"
	"log"
	"net"
	"os"
	"testing"
	"time"

	"hashserver/pkg/hashcalc"

	"github.com/pkg/errors"
	"github.com/pressly/goose"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

var MainConfig string = `
server:
  name: Compute hash server
  host: 127.0.0.1
  port: 8080
  worker-count: 5
  cache-count: 5
  db:
    # Supported DB type is postgres only
    type: postgres
    pool-count: 5
    host: localhost
    port: 5432
    user: postgres
    pass: postgres
    dbname: hashdb
    ssl: disable
metric:
  host: 127.0.0.1
  port: 7755
  path: metrics
logging:
  provider: graylog
  host: localhost
  port: 12201
  # Panic = 0
  # Fatal = 1
  # Error = 2
  # Warn = 3
  # Info = 4
  # Debug = 5 
  # Trace = 6
  level: 6
`

func makeConfig() error {
	hfile, err := os.OpenFile("test-config.yml", os.O_CREATE|os.O_RDWR, os.ModePerm)
	if err != nil {
		return err
	}
	_, err = hfile.WriteString(MainConfig)
	if err != nil {
		return err
	}
	hfile.Close()
	return nil
}

func ServerUp(ctx context.Context, cfg *config.Config) {
	dbPool, err := psql.New(cfg.Server.DSN, cfg.Server.DB.PoolCount)
	if err != nil {
		log.Fatal(err)
	}

	s := grpc.NewServer()
	server := &handlers.Server{}
	server.DB = dbPool
	server.Logger = MyLogger
	server.Cache = LRUCache.NewLRUCache(cfg.Server.CacheCount)
	server.Workers = cfg.Server.HashWorkers
	hashcalc.RegisterHashCalcServer(s, server)
	listenAddr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	lis, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Fatal(err)
	}

	for {
		select {
		case <-ctx.Done():
			s.GracefulStop()
			fmt.Println("close the server")
			return
		default:
			fmt.Printf("Starting server %s", listenAddr)
			if err := s.Serve(lis); err != nil {
				log.Fatal(err)
			}
		}
	}

}

func TestMain(t *testing.T) {
	var err error
	req := require.New(t)
	makeConfig()
	defer os.Remove("test-config.yml")

	cfg, err = config.NewConfig("test-config.yml")
	if err != nil {
		log.Fatal(err)
	}

	logger := logrus.New()
	logger.SetOutput(os.Stdout)
	logger.SetLevel(logrus.InfoLevel)
	MyLogger = logger

	ms := metrics.NewMetricServer()
	ms.Host = cfg.Metrics.Host
	ms.Port = cfg.Metrics.Port
	ms.Path = cfg.Metrics.Path
	ms.MetricLog = MyLogger
	go ms.Start()

	mdb, err := sql.Open("postgres", cfg.Server.DSN)
	if err != nil {
		log.Fatal(errors.WithStack(err))
	}
	err = mdb.Ping()
	if err != nil {
		log.Fatal(errors.WithStack(err))
	}
	defer mdb.Close()
	if err := goose.Up(mdb, "../migrations"); err != nil {
		log.Fatal(errors.WithStack(err))
	}
	defer goose.Down(mdb, "../migrations")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go ServerUp(ctx, cfg)
	time.Sleep(3 * time.Second)

	grpServer := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	cwt, _ := context.WithTimeout(context.Background(), time.Second*1)
	conn, err := grpc.DialContext(cwt, grpServer, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		MyLogger.Errorf("%+v\n", err)
		return
	}
	SHA3Calc := hashcalc.NewHashCalcClient(conn)
	var toServer hashcalc.StringList
	toServer.Lines = []string{"line1"}
	result, err := SHA3Calc.ComputeHash(context.TODO(), &toServer)
	if err != nil {
		MyLogger.Errorf("%+v\n", err)
		return
	}
	req.Equal(1, len(result.Hash))

	output, err := SHA3Calc.GetHash(context.TODO(), &hashcalc.IDList{Ids: []int64{result.Hash[0].Id}})
	if err != nil {
		MyLogger.Errorf("%+v\n", err)
		return
	}
	req.Equal(result.Hash[0].Hash, output.Hash[0].Hash)

	toServer.Lines = []string{"line2"}
	result, err = SHA3Calc.ComputeHash(context.TODO(), &toServer)
	if err != nil {
		MyLogger.Errorf("%+v\n", err)
		return
	}
	req.Equal(1, len(result.Hash))
	output, err = SHA3Calc.GetHash(context.TODO(), &hashcalc.IDList{Ids: []int64{result.Hash[0].Id}})
	if err != nil {
		MyLogger.Errorf("%+v\n", err)
		return
	}
	req.Equal(result.Hash[0].Hash, output.Hash[0].Hash)
}
