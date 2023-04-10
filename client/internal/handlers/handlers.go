package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"hash-service-client/pkg/hashcalc"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"hash-service-client/internal/config"
	metrics "hash-service-client/internal/metrics"
	"hash-service-client/internal/middleware"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

const (
	timeout time.Duration = 5
)

func (h *Handler) Send(w http.ResponseWriter, r *http.Request) {
	metrics.SendCall.Add(1)
	UUID := r.Context().Value(middleware.RequestContextID).(string)
	logger := r.Context().Value("logger").(*logrus.Logger)
	cfg := r.Context().Value("config").(*config.Config)

	dataBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"service":  "WEBUI",
			"module":   "client",
			"uuid":     UUID,
			"function": "Send",
			"error":    fmt.Sprintf("%v", errors.WithStack(err)),
		}).Error("not able to read request body")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	r.Body.Close()

	params, _ := url.ParseQuery(string(dataBody))

	var ok bool
	var toServer hashcalc.StringList
	toServer.Uuid = UUID
	if toServer.Lines, ok = params["params"]; !ok {
		logger.WithFields(logrus.Fields{
			"service":  "WEBUI",
			"module":   "client",
			"uuid":     UUID,
			"function": "Send",
			"error":    fmt.Sprintf("%v", errors.WithStack(err)),
		}).Error("no input data")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	grpServer := fmt.Sprintf("%s:%s", cfg.Grpc.Host, cfg.Grpc.Port)
	cwt, _ := context.WithTimeout(context.Background(), time.Second*timeout)
	conn, err := grpc.DialContext(cwt, grpServer, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		logger.WithFields(logrus.Fields{
			"service":           "WEBUI",
			"module":            "client",
			"uuid":              UUID,
			"function":          "Send",
			"error":             fmt.Sprintf("%v", errors.WithStack(err)),
			"connection string": grpServer,
		}).Error("gRPC server error")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer conn.Close()
	logger.WithFields(logrus.Fields{
		"service":  "WEBUI",
		"module":   "client",
		"uuid":     UUID,
		"function": "Send",
	}).Info("connected to GRPC")

	SHA3Calc := hashcalc.NewHashCalcClient(conn)
	result, err := SHA3Calc.ComputeHash(context.TODO(), &toServer)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"service":  "WEBUI",
			"module":   "client",
			"uuid":     UUID,
			"function": "Send",
			"error":    fmt.Sprintf("%v", errors.WithStack(err)),
		}).Error("error to compute values")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if len(result.Hash) == 0 {
		logger.WithFields(logrus.Fields{
			"service":  "WEBUI",
			"module":   "client",
			"uuid":     UUID,
			"function": "Send",
			"error":    "no data from server",
		}).Error("empty response")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	logger.WithFields(logrus.Fields{
		"service":  "WEBUI",
		"module":   "client",
		"uuid":     UUID,
		"function": "Send",
	}).Info("received data from gRPC server")

	logger.WithFields(logrus.Fields{
		"service":  "WEBUI",
		"module":   "client",
		"uuid":     UUID,
		"function": "Send",
	}).Trace(fmt.Sprintf("payload: %+v", result.Hash))

	data, err := json.Marshal(result.Hash)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"service":  "WEBUI",
			"module":   "client",
			"uuid":     UUID,
			"function": "Send",
			"error":    fmt.Sprintf("%v", errors.WithStack(err)),
		}).Error("marshalling error")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	logger.WithFields(logrus.Fields{
		"service":  "WEBUI",
		"module":   "client",
		"uuid":     UUID,
		"function": "Send",
	}).Debug("send response to client")

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)

}

func (h *Handler) Check(w http.ResponseWriter, r *http.Request) {
	metrics.CheckCall.Add(1)
	UUID := r.Context().Value(middleware.RequestContextID).(string)
	logger := r.Context().Value("logger").(*logrus.Logger)
	cfg := r.Context().Value("config").(*config.Config)

	inputStr := r.URL.Query().Get("ids")
	input := strings.Split(inputStr, ",")
	var getHash hashcalc.IDList
	getHash.Ids = make([]int64, 0, len(input))

	for i := 0; i < len(input); i++ {
		number, err := strconv.Atoi(input[i])
		if err != nil {
			logger.WithFields(logrus.Fields{
				"service":  "WEBUI",
				"module":   "client",
				"uuid":     UUID,
				"function": "Check",
				"error":    fmt.Sprintf("%v", errors.WithStack(err)),
			}).Error("problem to convert data to integer")
			continue
		}
		getHash.Ids = append(getHash.Ids, int64(number))
	}

	if len(getHash.Ids) == 0 {
		log.Printf("Input data is empty")
		logger.WithFields(logrus.Fields{
			"service":  "WEBUI",
			"module":   "client",
			"uuid":     UUID,
			"function": "Check",
			"error":    "no data",
		}).Error("request is empty")

		w.WriteHeader(http.StatusBadRequest)
		return
	}

	grpServer := fmt.Sprintf("%s:%s", cfg.Grpc.Host, cfg.Grpc.Port)
	cwt, _ := context.WithTimeout(context.Background(), time.Second*timeout)
	conn, err := grpc.DialContext(cwt, grpServer, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		logger.WithFields(logrus.Fields{
			"service":           "WEBUI",
			"module":            "client",
			"uuid":              UUID,
			"function":          "Check",
			"error":             fmt.Sprintf("%v", err.Error()),
			"connection string": grpServer,
		}).Error("gRPC server error")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer conn.Close()
	logger.WithFields(logrus.Fields{
		"service":  "WEBUI",
		"module":   "client",
		"uuid":     UUID,
		"function": "Check",
	}).Info("connected to GRPC")

	SHA3Calc := hashcalc.NewHashCalcClient(conn)
	result, err := SHA3Calc.GetHash(context.TODO(), &getHash)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"service":  "WEBUI",
			"module":   "client",
			"uuid":     UUID,
			"function": "Check",
			"error":    fmt.Sprintf("%v", errors.WithStack(err)),
		}).Error("problem to get values")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if len(result.Hash) == 0 {
		logger.WithFields(logrus.Fields{
			"service":  "WEBUI",
			"module":   "client",
			"uuid":     UUID,
			"function": "Check",
			"error":    "no data",
		}).Info("no data from server")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	logger.WithFields(logrus.Fields{
		"service":  "WEBUI",
		"module":   "client",
		"uuid":     UUID,
		"function": "Check",
	}).Info("received data from gRPC server")

	data, err := json.Marshal(result.Hash)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"service":  "WEBUI",
			"module":   "client",
			"uuid":     UUID,
			"function": "Check",
			"error":    fmt.Sprintf("%v", errors.WithStack(err)),
		}).Error("marshalling error")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	logger.WithFields(logrus.Fields{
		"service":  "WEBUI",
		"module":   "client",
		"uuid":     UUID,
		"function": "Check",
	}).Debug("send response to client")

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)

}
