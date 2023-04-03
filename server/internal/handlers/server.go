package handlers

import (
	"context"
	b64 "encoding/base64"
	"fmt"
	LRUCache "hashserver/internal/cache"
	psql "hashserver/internal/database"
	"hashserver/pkg/hashcalc"
	"hashserver/pkg/hasher"

	"github.com/sirupsen/logrus"
)

type Server struct {
	DB     psql.Repository
	Logger *logrus.Logger
	Cache  *LRUCache.LRUCache
	hashcalc.UnimplementedHashCalcServer
}

func (s *Server) ComputeHash(ctx context.Context, input *hashcalc.StringList) (*hashcalc.ArrayHash, error) {
	var result hashcalc.ArrayHash
	result.Hash = make([]*hashcalc.Hash, 0, len(input.Lines))
	s.Logger.WithFields(logrus.Fields{
		"service":    "ComputeHash",
		"count":      len(input.Lines),
		"input data": input.Lines,
	}).Debug("start to compute")
	conn := s.DB.GetConnection()

	for i := 0; i < len(input.Lines); i++ {
		hash := hasher.Hash(input.Lines[i])
		s.Logger.WithFields(logrus.Fields{
			"service": "ComputeHash",
			"action":  "DB INSERT",
			"text":    input.Lines[i],
		}).Debug("hash calculated")

		b64string := b64.StdEncoding.EncodeToString([]byte(input.Lines[i]))
		id, err := psql.Insert(context.TODO(), conn, b64string, hash)
		if err != nil && err != psql.EmptyInsert {
			s.Logger.WithFields(logrus.Fields{
				"service": "ComputeHash",
				"action":  "DB INSERT",
				"text":    input.Lines[i],
			}).Errorf("failed to insert to DB: %v", err)
			return nil, err
		}
		if err == psql.EmptyInsert {
			s.Logger.WithFields(logrus.Fields{
				"service":    "ComputeHash",
				"action":     "DB INSERT",
				"input line": input.Lines[i],
			}).Infof("duplicate - skip")
			continue
		}

		s.Logger.WithFields(logrus.Fields{
			"service":    "ComputeHash",
			"id":         id,
			"input line": input.Lines[i],
		}).Info("Hash computed and inserted into DB")

		h := hashcalc.Hash{
			Id:   id,
			Hash: hash,
		}
		result.Hash = append(result.Hash, &h)
	}
	return &result, nil
}

func (s *Server) GetHash(ctx context.Context, in *hashcalc.IDList) (*hashcalc.ArrayHash, error) {
	conn := s.DB.GetConnection()

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

		hash, err := psql.Select(context.TODO(), conn, in.Ids[i])
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
