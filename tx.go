package sqlw

import (
	"context"
	"database/sql"
	"reflect"
	"sync"
)

type Tx struct {
	*sql.Tx
	mapping *sync.Map
	parser  func(field *reflect.StructField) string
}

func (tx *Tx) QueryRowContext(ctx context.Context, dst interface{}, query string, args ...interface{}) error {
	return queryRowContext(ctx, tx.Tx, tx.parser, dst, tx.mapping, query, args...)
}

func (tx *Tx) QueryRow(dst interface{}, query string, args ...interface{}) error {
	return tx.QueryRowContext(context.Background(), dst, query, args...)
}

func (tx *Tx) QueryContext(ctx context.Context, dst interface{}, query string, args ...interface{}) error {
	return queryContext(ctx, tx.Tx, tx.parser, dst, tx.mapping, query, args...)
}

func (tx *Tx) Query(dst interface{}, query string, args ...interface{}) error {
	return tx.QueryContext(context.Background(), dst, query, args...)
}
