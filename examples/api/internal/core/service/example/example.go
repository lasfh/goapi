package example

import (
	"context"

	"github.com/lasfh/goapi/examples/api/internal/resperr"
)

type exampleRepo interface {
	Find(ctx context.Context) ([]string, error)
}

type service struct {
	exexampleRepo exampleRepo
}

func NewService(
	exampleRepo exampleRepo,
) *service {
	return &service{
		exexampleRepo: exampleRepo,
	}
}

func (s *service) Find(ctx context.Context) ([]string, error) {
	list, err := s.exexampleRepo.Find(ctx)
	if err != nil {
		return nil, resperr.ErrInternalError.WithErrorContext(ctx, err)
	}

	return list, nil
}
