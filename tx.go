// Copyright 2022 lesismal. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package sqlw

import (
	"context"
	"database/sql"
)

type Tx struct {
	*DB
	*sql.Tx
}

func (tx *Tx) ExecContext(ctx context.Context, query string, args ...interface{}) (Result, error) {
	result, err := tx.Tx.ExecContext(ctx, query, args...)
	return newResult(result, query, args), err
}

func (tx *Tx) Exec(query string, args ...interface{}) (Result, error) {
	return tx.ExecContext(tx.ctx, query, args...)
}

func (tx *Tx) QueryRowContext(ctx context.Context, dst interface{}, query string, args ...interface{}) (Result, error) {
	return queryRowContext(ctx, tx.Tx, tx.parseFieldName, dst, tx.mapping, tx.rawScan, query, args...)
}

func (tx *Tx) QueryRow(dst interface{}, query string, args ...interface{}) (Result, error) {
	return tx.QueryRowContext(tx.ctx, dst, query, args...)
}

func (tx *Tx) QueryContext(ctx context.Context, dst interface{}, query string, args ...interface{}) (Result, error) {
	return queryContext(ctx, tx.Tx, tx.parseFieldName, dst, tx.mapping, tx.rawScan, query, args...)
}

func (tx *Tx) Query(dst interface{}, query string, args ...interface{}) (Result, error) {
	return tx.QueryContext(tx.ctx, dst, query, args...)
}

func (tx *Tx) SelectContext(ctx context.Context, dst interface{}, query string, args ...interface{}) (Result, error) {
	return tx.QueryContext(ctx, dst, query, args...)
}

func (tx *Tx) Select(dst interface{}, query string, args ...interface{}) (Result, error) {
	return tx.QueryContext(tx.ctx, dst, query, args...)
}

// deprecated.
// func (tx *Tx) SelectOneContext(ctx context.Context, dst interface{}, query string, args ...interface{}) (Result, error) {
// 	typ := reflect.TypeOf(dst)
// 	if !isStructPtr(typ) {
// 		return newResult(nil, query, args), fmt.Errorf("[sqlw %v] invalid dest type: %v", opTypSelect, typ)
// 	}
// 	return tx.SelectContext(tx.ctx , dst, query, args...)
// }

// deprecated.
// func (tx *Tx) SelectOne(dst interface{}, query string, args ...interface{}) (Result, error) {
// 	return tx.SelectOneContext(tx.ctx , dst, query, args...)
// }

func (tx *Tx) InsertContext(ctx context.Context, sqlHead string, args ...interface{}) (Result, error) {
	return insertContext(ctx, tx.Tx, nil, sqlHead, tx.DB, args...)
}

func (tx *Tx) Insert(sqlHead string, args ...interface{}) (Result, error) {
	return tx.InsertContext(tx.ctx, sqlHead, args...)
}

func (tx *Tx) UpdateContext(ctx context.Context, sqlHead string, args ...interface{}) (Result, error) {
	return updateByExecContext(ctx, tx.Tx, tx.DB, nil, sqlHead, args...)
}

func (tx *Tx) Update(sqlHead string, args ...interface{}) (Result, error) {
	return tx.UpdateContext(tx.ctx, sqlHead, args...)
}

func (tx *Tx) DeleteContext(ctx context.Context, query string, args ...interface{}) (Result, error) {
	result, err := tx.Tx.ExecContext(ctx, query, args...)
	return newResult(result, query, args), err
}

func (tx *Tx) Delete(query string, args ...interface{}) (Result, error) {
	return tx.DeleteContext(tx.ctx, query, args...)
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
