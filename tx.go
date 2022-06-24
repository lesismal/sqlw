package sqlw

import (
	"context"
	"database/sql"
)

type Tx struct {
	*DB
	*sql.Tx
}

func (tx *Tx) QueryRowContext(ctx context.Context, dst interface{}, query string, args ...interface{}) error {
	return queryRowContext(ctx, tx.Tx, tx.parseFieldName, dst, tx.mapping, query, args...)
}

func (tx *Tx) QueryRow(dst interface{}, query string, args ...interface{}) error {
	return tx.QueryRowContext(context.Background(), dst, query, args...)
}

func (tx *Tx) QueryContext(ctx context.Context, dst interface{}, query string, args ...interface{}) error {
	return queryContext(ctx, tx.Tx, tx.parseFieldName, dst, tx.mapping, query, args...)
}

func (tx *Tx) Query(dst interface{}, query string, args ...interface{}) error {
	return tx.QueryContext(context.Background(), dst, query, args...)
}

func (tx *Tx) InsertContext(ctx context.Context, sqlHead string, data interface{}) (sql.Result, error) {
	return insertContext(ctx, tx.DB.DB, nil, sqlHead, nil, data, tx.parseFieldName, tx.mapping)
}

func (tx *Tx) Insert(sqlHead string, data interface{}) (sql.Result, error) {
	return tx.InsertContext(context.Background(), sqlHead, data)
}
