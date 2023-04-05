package handlers

import (
	"context"

	"fmt"
	LRUCache "hashserver/internal/cache"
	psql "hashserver/internal/database"
	"hashserver/pkg/hashcalc"
	"sync"

	metrics "hashserver/internal/metrics"

	"github.com/pkg/errors"
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
	metrics.SendCall.Add(1)
	var result hashcalc.ArrayHash
	s.Logger.WithFields(logrus.Fields{
		"service": "ComputeHash",
		"module":  "server",
		"uuid":    input.Uuid,
		"count":   len(input.Lines),
	}).Info("start to compute")

	var wg sync.WaitGroup
	wg.Add(s.Workers)
	jobs, answer := workerInit(&wg, s.Workers)
	result.Hash = workPool(input.Lines, jobs, answer)
	wg.Wait()
	s.Logger.WithFields(logrus.Fields{
		"service": "ComputeHash",
		"module":  "server",
		"uuid":    input.Uuid,
	}).Debug("hashes calculated")

	err := s.DB.MultiHashInsert(context.TODO(), result.Hash)
	if err != nil {
		s.Logger.WithFields(logrus.Fields{
			"service":  "ComputeHash",
			"module":   "server",
			"uuid":     input.Uuid,
			"function": "MultiHashInsert",
			"error":    errors.WithStack(err),
		}).Error("not able to insert data into DB")
		return nil, fmt.Errorf("[ComputeHash] error %v", errors.WithStack(err))
	}

	s.Logger.WithFields(logrus.Fields{
		"service": "ComputeHash",
		"module":  "server",
		"uuid":    input.Uuid,
	}).Debug("hashes inserted into db")
	s.Logger.WithFields(logrus.Fields{
		"service": "ComputeHash",
		"module":  "server",
		"uuid":    input.Uuid,
		"count":   len(input.Lines),
	}).Info("finish to compute")

	return &result, nil
}

func (s *Server) GetHash(ctx context.Context, in *hashcalc.IDList) (*hashcalc.ArrayHash, error) {
	metrics.CheckCall.Add(1)
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
				"module":  "server",
				"uuid":    in.Uuid,
				"id":      in.Ids[i],
			}).Debug("found in cache")
			continue
		}
		hash, err := s.DB.Select(context.TODO(), in.Ids[i])
		if err != nil && err != psql.EmptySelect {
			s.Logger.WithFields(logrus.Fields{
				"service": "GetHash",
				"module":  "server",
				"uuid":    in.Uuid,
				"error":   errors.WithStack(err),
			}).Error("cannot select data from db")
			return nil, fmt.Errorf("[GetHash] selecting error %v", errors.WithStack(err))
		}
		if err == psql.EmptySelect {
			s.Logger.WithFields(logrus.Fields{
				"service": "GetHash",
				"module":  "server",
				"uuid":    in.Uuid,
				"id":      in.Ids[i],
			}).Infof("Request ID has not been found")
			continue
		}

		s.Logger.WithFields(logrus.Fields{
			"service": "GetHash",
			"module":  "server",
			"uuid":    in.Uuid,
			"id":      in.Ids[i],
		}).Debug("Hash fetched from DB")

		h := hashcalc.Hash{
			Id:   in.Ids[i],
			Hash: hash,
		}
		result.Hash = append(result.Hash, &h)
		s.Cache.Add(in.Ids[i], hash)
		s.Logger.WithFields(logrus.Fields{
			"service": "GetHash",
			"module":  "server",
			"uuid":    in.Uuid,
			"id":      in.Ids[i],
		}).Debug("added to cache")
	}
	return &result, nil
}
