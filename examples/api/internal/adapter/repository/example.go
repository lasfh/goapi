package repository

import (
	"context"

	"github.com/lasfh/goapi/database/gorm/repo"
	"gorm.io/gorm"
)

type exampleRepo struct {
	*repo.BaseRepository
}

func NewExampleRepo(db *gorm.DB) *exampleRepo {
	return &exampleRepo{
		BaseRepository: repo.NewBaseRepository(db),
	}
}

func (r *exampleRepo) Find(ctx context.Context) ([]string, error) {
	return []string{"A", "B", "C"}, nil
}
