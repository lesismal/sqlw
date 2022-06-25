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
	return queryRowContext(ctx, tx.Tx, tx.parseFieldName, dst, tx.mapping, tx.rawScan, query, args...)
}

func (tx *Tx) QueryRow(dst interface{}, query string, args ...interface{}) error {
	return tx.QueryRowContext(context.Background(), dst, query, args...)
}

func (tx *Tx) QueryContext(ctx context.Context, dst interface{}, query string, args ...interface{}) error {
	return queryContext(ctx, tx.Tx, tx.parseFieldName, dst, tx.mapping, tx.rawScan, query, args...)
}

func (tx *Tx) Query(dst interface{}, query string, args ...interface{}) error {
	return tx.QueryContext(context.Background(), dst, query, args...)
}

func (tx *Tx) InsertContext(ctx context.Context, sqlHead string, data interface{}) (Result, error) {
	return insertContext(ctx, tx.Tx, nil, sqlHead, data, tx.parseFieldName, tx.mapping)
}

func (tx *Tx) Insert(sqlHead string, data interface{}) (Result, error) {
	return tx.InsertContext(context.Background(), sqlHead, data)
}

func (tx *Tx) UpdateContext(ctx context.Context, sqlHead string, args ...interface{}) (Result, error) {
	return updateByExecContext(ctx, tx.Tx, nil, tx.parseFieldName, tx.mapping, sqlHead, args...)
}

func (tx *Tx) Update(sqlHead string, args ...interface{}) (Result, error) {
	return tx.UpdateContext(context.Background(), sqlHead, args...)
}

func (tx *Tx) PrepareContext(ctx context.Context, query string) (*Stmt, error) {
	stmt, err := tx.Tx.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}
	return NewStmt(tx.DB, stmt, query), nil
}

func (tx *Tx) Prepare(query string) (*Stmt, error) {
	stmt, err := tx.Tx.Prepare(query)
	if err != nil {
		return nil, err
	}
	return NewStmt(tx.DB, stmt, query), nil
}

func (tx *Tx) StmtContext(ctx context.Context, stmt *Stmt) *Stmt {
	return NewStmt(tx.DB, tx.Tx.StmtContext(ctx, stmt.Stmt), stmt.query)
}

func (tx *Tx) Stmt(stmt *Stmt) *Stmt {
	return NewStmt(tx.DB, tx.Tx.Stmt(stmt.Stmt), stmt.query)
}
