package handlers

import (
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWorker(t *testing.T) {
	var wg sync.WaitGroup
	workers := 1
	bigVal := strings.Repeat("c89082d423c04a9404b63c62ee45539345472e2ab6af2b01da85f12eed0faa2bcb26b4682fe860f0b26ddf615bef0962937daa4e5a1f0e0dd94685f3ab3e0179", 100)
	testData := []string{"123", "456", "789", "abc", "bcd", "def", bigVal}
	testResult := []string{
		"48c8947f69c054a5caa934674ce8881d02bb18fb59d5a63eeaddff735b0e9801e87294783281ae49fc8287a0fd86779b27d7972d3e84f0fa0d826d7cb67dfefc",
		"0fe220a126aeb06ab687b5cf73175abbd6194f57b593059f33186d72066a283af765cbbea04cae0bce0ce793116a4ac99424c28ea7fded4e88a18cfc51513cd4",
		"2ea57b140295dd51f2318778831c539e466bebcb93b82fe369fa3b2fe013de15d7a364737915451f674eac258e43e9892e7ddc5f6f319b104be517feaaf6aa3c",
		"b751850b1a57168a5693cd924b6b096e08f621827444f70d884f5d0240d2712e10e116e9192af3c91a7ec57647e3934057340b4cf408d5a56592f8274eec53f0",
		"d076b21ef66d2acefc7e17cc8f5115e70a6d39cd324467c2b638b43d601d163b149530922598613b72cea9f749304395a0e9b0900540aa225c9f11d8207f3a03",
		"ebc4090f9fec594fabda0b5fb87f4ade83630ef428af56677dc598debec9d5ee8252928001d304c6d868b7b7f4ee99320ee761988509c142177ac687e7c3b198",
		"103da31125855596e4a30fcde41a01ba2b1d022d5912fb0f4fa2b38246c9d4de4206fcd3b32f3d7d352fc7e5d72a1b3d3d65007aa28a0f0b9937a4cc47a24018",
	}
	wg.Add(workers)
	jobs, answer := workerInit(&wg, workers, len(testData))
	result := workPool(testData, jobs, answer)
	wg.Wait()
	req := require.New(t)
	match := 0
	for _, expected := range testResult {
		for _, actual := range result {
			if expected == actual.Hash {
				match++
				continue
			}
		}
	}
	req.Equal(7, match)
}
