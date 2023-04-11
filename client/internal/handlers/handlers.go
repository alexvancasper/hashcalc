package handlers

import (
	"encoding/json"
	"fmt"
	"hash-service-client/pkg/hashcalc"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	metrics "hash-service-client/internal/metrics"
	"hash-service-client/internal/middleware"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/metadata"
)

func (h *Handler) Send(w http.ResponseWriter, r *http.Request) {
	metrics.SendCall.Add(1)
	UUID := r.Context().Value(middleware.RequestContextID).(string)

	l := h.Logger.WithFields(logrus.Fields{
		"service":  "WEBUI",
		"module":   "client",
		"uuid":     UUID,
		"function": "Send",
	})
	l.Info("Start to process the request")

	dataBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		l.WithFields(logrus.Fields{
			"error": fmt.Sprintf("%v", errors.WithStack(err)),
		}).Error("not able to read request body")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	params, _ := url.ParseQuery(string(dataBody)) //TODO: need to refactor it

	var ok bool
	var toServer hashcalc.StringList
	if toServer.Lines, ok = params["params"]; !ok {
		l.WithField("error", fmt.Sprintf("%v", errors.WithStack(err))).Error("no input data")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	SHA3Calc := hashcalc.NewHashCalcClient(h.Server)
	header := metadata.New(map[string]string{"X-REQUEST-ID": UUID})
	ctx := metadata.NewOutgoingContext(r.Context(), header)

	result, err := SHA3Calc.ComputeHash(ctx, &toServer)
	if err != nil {
		l.WithField("error", fmt.Sprintf("%v", errors.WithStack(err))).Error("error to compute values")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if len(result.Hash) == 0 {
		l.WithField("error", fmt.Sprintf("%v", errors.WithStack(err))).Error("empty response")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	l.Info("received data from gRPC server")
	l.Trace(fmt.Sprintf("payload: %+v", result.Hash))

	data, err := json.Marshal(result.Hash)
	if err != nil {
		l.WithField("error", fmt.Sprintf("%v", errors.WithStack(err))).Error("marshalling error")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	l.Debug("send response to client")

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)

}

func (h *Handler) Check(w http.ResponseWriter, r *http.Request) {
	metrics.CheckCall.Add(1)
	UUID := r.Context().Value(middleware.RequestContextID).(string)
	l := h.Logger.WithFields(logrus.Fields{
		"service":  "WEBUI",
		"module":   "client",
		"uuid":     UUID,
		"function": "Check",
	})

	l.Info("Start to process the request")
	inputStr := r.URL.Query().Get("ids")
	input := strings.Split(inputStr, ",")
	var getHash hashcalc.IDList
	getHash.Ids = make([]int64, 0, len(input))

	for i := 0; i < len(input); i++ {
		number, err := strconv.Atoi(input[i])
		if err != nil {
			l.WithField("error", fmt.Sprintf("%v", errors.WithStack(err))).Error("problem to convert data to integer")
			continue
		}
		getHash.Ids = append(getHash.Ids, int64(number))
	}

	if len(getHash.Ids) == 0 {
		l.WithField("error", "no content").Error("request is empty")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	SHA3Calc := hashcalc.NewHashCalcClient(h.Server)
	header := metadata.New(map[string]string{"X-REQUEST-ID": UUID})
	ctx := metadata.NewOutgoingContext(r.Context(), header)

	result, err := SHA3Calc.GetHash(ctx, &getHash)
	if err != nil {
		l.WithField("error", fmt.Sprintf("%v", errors.WithStack(err))).Error("problem to get values")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if len(result.Hash) == 0 {
		l.WithField("error", "no content").Info("no data from server")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	l.Info("received data from gRPC server")
	l.Trace(fmt.Sprintf("payload: %+v", result.Hash))

	data, err := json.Marshal(result.Hash)
	if err != nil {
		l.WithField("error", fmt.Sprintf("%v", errors.WithStack(err))).Error("marshalling error")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	l.Debug("send response to client")

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}
