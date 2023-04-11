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
	"google.golang.org/grpc/metadata"
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
	var uuid string
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		uuid = md.Get("X-REQUEST-ID")[0]
	}

	l := s.Logger.WithFields(logrus.Fields{
		"service":  "gRPC server",
		"module":   "server",
		"uuid":     uuid,
		"function": "ComputeHash",
	})

	l.WithField("count", len(input.Lines)).Info("start to compute")

	var wg sync.WaitGroup
	wg.Add(s.Workers)
	jobs, answer := workerInit(&wg, s.Workers, len(input.Lines))
	result.Hash = workPool(input.Lines, jobs, answer)
	wg.Wait()
	l.Debug("hashes calculated")

	err := s.DB.MultiHashInsert(context.TODO(), result.Hash)
	if err != nil {
		l.WithField("error", fmt.Sprintf("%v", errors.WithStack(err))).Error("not able to insert data into DB")
		return nil, fmt.Errorf("[ComputeHash] error %v", errors.WithStack(err))
	}
	l.Debug("hashes inserted into db")
	l.Info("finish to compute")
	l.Trace(fmt.Sprintf("payload: %+v", result.Hash))

	return &result, nil
}

func (s *Server) GetHash(ctx context.Context, in *hashcalc.IDList) (*hashcalc.ArrayHash, error) {
	metrics.CheckCall.Add(1)
	var result hashcalc.ArrayHash
	var uuid string
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		uuid = md.Get("X-REQUEST-ID")[0]
	}
	l := s.Logger.WithFields(logrus.Fields{
		"service":  "gRPC server",
		"module":   "server",
		"uuid":     uuid,
		"function": "GetHash",
	})

	result.Hash = make([]*hashcalc.Hash, 0, len(in.Ids))

	l.Info("start to fetch data")

	for i := 0; i < len(in.Ids); i++ {

		if value, ok := s.Cache.Get(in.Ids[i]); ok {
			result.Hash = append(result.Hash, &hashcalc.Hash{
				Id:   in.Ids[i],
				Hash: value,
			})
			l.WithField("id", in.Ids[i]).Debug("found in cache")
			continue
		}
		hash, err := s.DB.Select(context.TODO(), in.Ids[i])
		if err != nil && err != psql.EmptySelect {
			l.WithField("error", fmt.Sprintf("%v", errors.WithStack(err))).Error("cannot select data from db")
			return nil, fmt.Errorf("[GetHash] selecting error %v", errors.WithStack(err))
		}
		if err == psql.EmptySelect {
			l.WithField("id", in.Ids[i]).Infof("Request ID has not been found")
			continue
		}

		l.WithField("id", in.Ids[i]).Debug("Hash fetched from DB")

		h := hashcalc.Hash{
			Id:   in.Ids[i],
			Hash: hash,
		}
		result.Hash = append(result.Hash, &h)
		s.Cache.Add(in.Ids[i], hash)
		l.WithField("id", in.Ids[i]).Debug("added to cache")
	}

	l.WithField("count", len(in.Ids)).Info("finish to fetch data")

	return &result, nil
}
