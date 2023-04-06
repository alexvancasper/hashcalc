package handlers

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWorker(t *testing.T) {
	var wg sync.WaitGroup
	workers := 2
	testData := []string{"123", "456", "789"}
	testResult := []string{
		"48c8947f69c054a5caa934674ce8881d02bb18fb59d5a63eeaddff735b0e9801e87294783281ae49fc8287a0fd86779b27d7972d3e84f0fa0d826d7cb67dfefc",
		"0fe220a126aeb06ab687b5cf73175abbd6194f57b593059f33186d72066a283af765cbbea04cae0bce0ce793116a4ac99424c28ea7fded4e88a18cfc51513cd4",
		"2ea57b140295dd51f2318778831c539e466bebcb93b82fe369fa3b2fe013de15d7a364737915451f674eac258e43e9892e7ddc5f6f319b104be517feaaf6aa3c",
	}
	wg.Add(workers)
	jobs, answer := workerInit(&wg, workers)
	result := workPool(testData, jobs, answer)
	wg.Wait()
	req := require.New(t)
	for i := 0; i < len(testResult); i++ {
		req.Equal(testResult[i], result[i].Hash)
	}
}
