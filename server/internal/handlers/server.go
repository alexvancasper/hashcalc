package handlers

import (
	"context"

	"fmt"
	LRUCache "hashserver/internal/cache"
	psql "hashserver/internal/database"
	"hashserver/pkg/hashcalc"
	"sync"

	"github.com/sirupsen/logrus"
)

type Server struct {
	DB      *psql.Instance
	Logger  *logrus.Logger
	Cache   *LRUCache.LRUCache
	Workers int
	hashcalc.UnimplementedHashCalcServer
}

func (s *Server) ComputeHash(ctx context.Context, input *hashcalc.StringList) (*hashcalc.ArrayHash, error) {
	var result hashcalc.ArrayHash
	s.Logger.WithFields(logrus.Fields{
		"service":    "ComputeHash",
		"count":      len(input.Lines),
		"input data": input.Lines,
	}).Debug("start to compute")

	var wg sync.WaitGroup
	wg.Add(s.Workers)
	jobs, answer := workerInit(&wg, s.Workers)
	result.Hash = workPool(input.Lines, jobs, answer)
	wg.Wait()
	err := s.DB.MultiHashInsert(context.TODO(), result.Hash)
	if err != nil {
		s.Logger.WithFields(logrus.Fields{
			"service":    "ComputeHash",
			"count":      len(input.Lines),
			"input data": input.Lines,
		}).Debug("start to compute")
		return nil, fmt.Errorf("[ComputeHash] error %v", err)
	}

	return &result, nil
}

func (s *Server) GetHash(ctx context.Context, in *hashcalc.IDList) (*hashcalc.ArrayHash, error) {
	var result hashcalc.ArrayHash
	result.Hash = make([]*hashcalc.Hash, 0, len(in.Ids))

	for i := 0; i < len(in.Ids); i++ {

		if value, ok := s.Cache.Get(in.Ids[i]); ok {
			result.Hash = append(result.Hash, &hashcalc.Hash{
				Id:   in.Ids[i],
				Hash: value,
			})
			s.Logger.WithFields(logrus.Fields{
				"service": "GetHash",
				"action":  "CACHE Check",
				"id":      in.Ids[i],
			}).Debug("Hash has been found in cache")
			continue
		}
		hash, err := s.DB.Select(context.TODO(), in.Ids[i])
		if err != nil && err != psql.EmptySelect {
			s.Logger.WithFields(logrus.Fields{
				"service": "GetHash",
				"action":  "DB SELECT",
			}).Error(err)
			return nil, fmt.Errorf("[GetHash] selecting error %v", err)
		}
		if err == psql.EmptySelect {
			s.Logger.WithFields(logrus.Fields{
				"service": "GetHash",
				"action":  "DB SELECT",
				"id":      in.Ids[i],
			}).Infof("Request ID has not been found ")
			continue
		}

		s.Logger.WithFields(logrus.Fields{
			"service": "GetHash",
			"id":      in.Ids[i],
		}).Info("Hash fetched from DB")

		h := hashcalc.Hash{
			Id:   in.Ids[i],
			Hash: hash,
		}
		result.Hash = append(result.Hash, &h)
		s.Cache.Add(in.Ids[i], hash)
		s.Logger.WithFields(logrus.Fields{
			"service": "GetHash",
			"action":  "CACHE Check",
			"id":      in.Ids[i],
		}).Debug("Hash has been added into cache")
	}
	return &result, nil
}
