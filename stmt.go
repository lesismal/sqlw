package sqlw

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
)

type Stmt struct {
	*DB
	*sql.Stmt
	query string
}

func (stmt *Stmt) Sql(ctx context.Context, dst interface{}, args ...interface{}) string {
	return stmt.query
}

func (stmt *Stmt) QueryRowContext(ctx context.Context, dst interface{}, args ...interface{}) (Result, error) {
	if dst == nil {
		return nil, fmt.Errorf("[sqlw %v] invalid dest value nil: %v", opTypSelect, reflect.TypeOf(dst))
	}

	rows, err := stmt.Stmt.QueryContext(ctx, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	err = rowsToStruct(rows, dst, stmt.parseFieldName, stmt.mapping, sqlMappingKey(opTypSelect, stmt.query, reflect.TypeOf(dst)), stmt.rawScan)
	if err != nil {
		return nil, err
	}
	return newResult(nil, stmt.query, args), nil
}

func (stmt *Stmt) QueryRow(dst interface{}, args ...interface{}) (Result, error) {
	return stmt.QueryRowContext(context.Background(), dst, args...)
}

func (stmt *Stmt) QueryContext(ctx context.Context, dst interface{}, args ...interface{}) (Result, error) {
	rows, err := stmt.Stmt.QueryContext(ctx, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if isStructPtr(reflect.TypeOf(dst)) {
		err = rowsToStruct(rows, dst, stmt.parseFieldName, stmt.mapping, sqlMappingKey(opTypSelect, stmt.query, reflect.TypeOf(dst)), stmt.rawScan)
		if err != nil {
			return nil, err
		}
		return newResult(nil, stmt.query, args), nil
	}

	err = rowsToSlice(rows, dst, stmt.parseFieldName, stmt.mapping, sqlMappingKey(opTypSelect, stmt.query, reflect.TypeOf(dst)), stmt.rawScan)
	if err != nil {
		return nil, err
	}
	return newResult(nil, stmt.query, args), nil
}

func (stmt *Stmt) Query(dst interface{}, args ...interface{}) (Result, error) {
	return stmt.QueryContext(context.Background(), dst, args...)
}

func (stmt *Stmt) SelectContext(ctx context.Context, dst interface{}, query string, args ...interface{}) (Result, error) {
	return stmt.QueryContext(ctx, dst, args...)
}

func (stmt *Stmt) Select(dst interface{}, query string, args ...interface{}) (Result, error) {
	return stmt.QueryContext(context.Background(), dst, args...)
}

func (stmt *Stmt) InsertContext(ctx context.Context, data ...interface{}) (Result, error) {
	return insertContext(ctx, nil, stmt, stmt.query, data, stmt.parseFieldName, stmt.mapping)
}

func (stmt *Stmt) Insert(data ...interface{}) (Result, error) {
	return stmt.InsertContext(context.Background(), data...)
}

func (stmt *Stmt) UpdateContext(ctx context.Context, args ...interface{}) (Result, error) {
	return updateByExecContext(ctx, nil, stmt, stmt.parseFieldName, stmt.mapping, stmt.query, args...)
}

func (stmt *Stmt) Update(args ...interface{}) (Result, error) {
	return stmt.UpdateContext(context.Background(), args...)
}

func NewStmt(db *DB, stmt *sql.Stmt, query string) *Stmt {
	return &Stmt{
		DB:    db,
		Stmt:  stmt,
		query: query,
	}
}
