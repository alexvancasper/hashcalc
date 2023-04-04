package handlers

import (
	"hashserver/pkg/hashcalc"
	"hashserver/pkg/hasher"
	"sync"
)

type workers struct {
	Number int
	jobs   chan<- hashcalc.Hash
	result <-chan hashcalc.Hash
}

func workerInit(wg *sync.WaitGroup, num int) (chan<- hashcalc.Hash, <-chan hashcalc.Hash) {
	jobs := make(chan hashcalc.Hash, num)
	result := make(chan hashcalc.Hash, num)
	for num > 0 {
		go worker(wg, num, jobs, result)
		num--
	}
	return jobs, result
}

func workPool(task []string, jobs chan<- hashcalc.Hash, result <-chan hashcalc.Hash) []*hashcalc.Hash {
	for j := 0; j < len(task); j++ {
		jobs <- hashcalc.Hash{
			Id:   int64(j),
			Hash: task[j],
		}
	}
	close(jobs)

	output := make([]*hashcalc.Hash, 0, len(task))
	for i := 0; i < len(task); i++ {
		oneHash := <-result
		output = append(output, &oneHash)
	}
	return output
}

func worker(wg *sync.WaitGroup, id int, job <-chan hashcalc.Hash, result chan<- hashcalc.Hash) {
	defer wg.Done()
	for j := range job {
		result <- hashcalc.Hash{
			Id:   int64(j.Id),
			Hash: hasher.Hash(j.Hash),
		}
	}
}
