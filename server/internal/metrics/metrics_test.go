package metrics

import (
	"fmt"
	"net/http"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestStar(t *testing.T) {
	req := require.New(t)

	logger := logrus.New()
	logger.SetOutput(os.Stdout)
	logger.SetLevel(logrus.InfoLevel)

	srv := NewMetricServer()
	srv.Host = "127.0.0.1"
	srv.Port = "7744"
	srv.Path = "test-metrics"
	srv.MetricLog = logger
	go srv.Start()
	time.Sleep(1 * time.Second)

	URL := fmt.Sprintf("http://%s:%s/%s", srv.Host, srv.Port, srv.Path)
	resp, err := http.Get(URL)
	if err != nil {
		t.Errorf("client get error %+v", err)
	}
	defer resp.Body.Close()
	stopSrv.server <- syscall.SIGINT
	req.Equal(http.StatusOK, resp.StatusCode)
}
