package unitofwork

import (
	"context"

	"gorm.io/gorm"
)

type txContextKey struct{}

type transaction struct {
	db *gorm.DB
}

func NewTransactionManager(db *gorm.DB) *transaction {
	return &transaction{db}
}

func (t *transaction) Execute(ctx context.Context, fn func(ctx context.Context) error) error {
	if tx := t.getTransactionFromContext(ctx); tx != nil {
		return fn(ctx)
	}

	return t.db.Transaction(func(tx *gorm.DB) error {
		ctx = context.WithValue(ctx, txContextKey{}, tx)

		return fn(ctx)
	})
}

func (t *transaction) getTransactionFromContext(ctx context.Context) *gorm.DB {
	if tx, ok := ctx.Value(txContextKey{}).(*gorm.DB); ok {
		return tx
	}
	return nil
}

func DB(ctx context.Context, db *gorm.DB) *gorm.DB {
	if tx, ok := ctx.Value(txContextKey{}).(*gorm.DB); ok {
		return tx
	}

	return db
}
