package main

import (
	"context"
	"fmt"
	"hash-service-client/pkg/hashcalc"
	"time"

	"google.golang.org/grpc"
)

func main() {

	cwt, _ := context.WithTimeout(context.Background(), time.Second*5)
	conn, err := grpc.DialContext(cwt, "localhost:50051", grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	SHA3Calc := hashcalc.NewHashCalcClient(conn)

	lines := []string{"line1", "line2"}

	var toServer hashcalc.TextList
	toServer.Lines = lines
	result, err := SHA3Calc.CalcSHA3(context.TODO(), &toServer)
	if err != nil {
		fmt.Printf("error: %v", err)
		return
	}

	fmt.Println(result)
}
