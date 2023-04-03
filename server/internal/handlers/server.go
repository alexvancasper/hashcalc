package handlers

import (
	"context"
	psql "hashserver/internal/database"
	"hashserver/pkg/hashcalc"
	"hashserver/pkg/hasher"

	"github.com/sirupsen/logrus"
)

type Server struct {
	DB     psql.Repository
	Logger *logrus.Logger
	hashcalc.UnimplementedHashCalcServer
}

func (s *Server) ComputeHash(ctx context.Context, input *hashcalc.StringList) (*hashcalc.ArrayHash, error) {
	var result hashcalc.ArrayHash
	result.Hash = make([]*hashcalc.Hash, len(input.Lines))
	for i := 0; i < len(input.Lines); i++ {
		hash := hasher.Hash(input.Lines[i])
		psql.Insert()
		h := hashcalc.Hash{
			Id:   int64(1),
			Hash: hash,
		}
		result.Hash = append(result.Hash, &h)
	}

	return &result, nil
}

func (s *Server) GetHash(ctx context.Context, in *hashcalc.IDList) (*hashcalc.ArrayHash, error) {
	return nil, nil
}
