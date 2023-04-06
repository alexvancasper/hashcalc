package config

import (
	"fmt"
	"os"
	"testing"
)

var myYaml string = `
server:
  name: Compute hash server
  host: 0.0.0.0
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
    dbname: testdb
    ssl: true
metric:
  host: 0.0.0.0
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

func TestNewConfig(t *testing.T) {
	hfile, err := os.OpenFile("test-config.yml", os.O_CREATE|os.O_RDWR, os.ModePerm)
	if err != nil {
		t.Fatal("cannot create test file")
	}
	_, err = hfile.WriteString(myYaml)
	if err != nil {
		t.Fatal("cannot write to test file")
	}
	hfile.Close()
	defer os.Remove("test-config.yml")

	cfg, err := NewConfig("test-config.yml")
	if err != nil {
		t.Fail()
	}
	fmt.Printf("%+v\n", cfg)
	if cfg.Metrics.Path != "metrics" {
		t.Errorf("%s != metrics", cfg.Metrics.Path)
	}
	if cfg.Server.DSN != "host=localhost port=5432 user=postgres password=postgres dbname=testdb sslmode=true" {
		t.Errorf("%s != host=localhost port=5432 user=postgres password=postgres dbname=testdb sslmode=true", cfg.Metrics.Path)
	}

}
