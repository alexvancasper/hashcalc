package handlers

import (
	"context"
	"encoding/json"
	"hash-service-client/pkg/hashcalc"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"google.golang.org/grpc"
)

const (
	grpServer string        = "localhost:8080"
	timeout   time.Duration = 5
)

func (h *Handler) Send(w http.ResponseWriter, r *http.Request) {
	dataBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error getting body of request %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	r.Body.Close()

	params, _ := url.ParseQuery(string(dataBody))

	var ok bool
	var toServer hashcalc.StringList
	if toServer.Lines, ok = params["line"]; !ok {
		log.Printf("no input data %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	cwt, _ := context.WithTimeout(context.Background(), time.Second*timeout)
	conn, err := grpc.DialContext(cwt, grpServer, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Printf("gRPC Server error %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer conn.Close()
	SHA3Calc := hashcalc.NewHashCalcClient(conn)
	result, err := SHA3Calc.ComputeHash(context.TODO(), &toServer)
	if err != nil {
		log.Printf("error to compute values %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(result.Hash) == 0 {
		log.Printf("no data recevied")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	data, err := json.Marshal(result.Hash)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)

}

func (h *Handler) Check(w http.ResponseWriter, r *http.Request) {
	inputStr := r.URL.Query().Get("ids")
	input := strings.Split(inputStr, ",")
	var getHash hashcalc.IDList
	getHash.Ids = make([]int64, 0, len(input))

	for i := 0; i < len(input); i++ {
		number, err := strconv.Atoi(input[i])
		if err != nil {
			log.Printf("error of converting value to integer %v", err)
			continue
		}
		getHash.Ids = append(getHash.Ids, int64(number))
	}

	if len(getHash.Ids) == 0 {
		log.Printf("Input data is empty")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	cwt, _ := context.WithTimeout(context.Background(), time.Second*timeout)
	conn, err := grpc.DialContext(cwt, grpServer, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Printf("gRPC Server error %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	SHA3Calc := hashcalc.NewHashCalcClient(conn)
	result, err := SHA3Calc.GetHash(context.TODO(), &getHash)
	if err != nil {
		log.Printf("error getting data from server %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if len(result.Hash) == 0 {
		log.Printf("no data on the server")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	data, err := json.Marshal(result.Hash)
	if err != nil {
		log.Printf("marshalling error %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)

}
