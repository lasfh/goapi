package repo

import (
	"context"

	unitofwork "github.com/lasfh/goapi/database/gorm/unit_of_work"
	"gorm.io/gorm"
)

type BaseRepository struct {
	DB func(context.Context) *gorm.DB
}

func NewBaseRepository(
	db *gorm.DB,
) *BaseRepository {
	return &BaseRepository{
		DB: func(ctx context.Context) *gorm.DB {
			return unitofwork.DB(ctx, db)
		},
	}
}
