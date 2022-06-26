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

func (stmt *Stmt) ExecContext(ctx context.Context, args ...interface{}) (Result, error) {
	result, err := stmt.Stmt.ExecContext(ctx, args...)
	return newResult(result, stmt.query, args), err
}

func (stmt *Stmt) Exec(args ...interface{}) (Result, error) {
	return stmt.ExecContext(context.Background(), args...)
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
	return newResult(nil, stmt.query, args), err
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
		return newResult(nil, stmt.query, args), err
	}

	err = rowsToSlice(rows, dst, stmt.parseFieldName, stmt.mapping, sqlMappingKey(opTypSelect, stmt.query, reflect.TypeOf(dst)), stmt.rawScan)
	return newResult(nil, stmt.query, args), err
}

func (stmt *Stmt) Query(dst interface{}, args ...interface{}) (Result, error) {
	return stmt.QueryContext(context.Background(), dst, args...)
}

func (stmt *Stmt) SelectContext(ctx context.Context, dst interface{}, args ...interface{}) (Result, error) {
	return stmt.QueryContext(ctx, dst, args...)
}

func (stmt *Stmt) Select(dst interface{}, args ...interface{}) (Result, error) {
	return stmt.QueryContext(context.Background(), dst, args...)
}

func (stmt *Stmt) SelectOneContext(ctx context.Context, dst interface{}, args ...interface{}) (Result, error) {
	typ := reflect.TypeOf(dst)
	if !isStructPtr(typ) {
		return newResult(nil, stmt.query, args), fmt.Errorf("[sqlw %v] invalid dest type: %v", opTypSelect, typ)
	}
	return stmt.SelectContext(ctx, dst, args...)
}

func (stmt *Stmt) SelectOne(dst interface{}, args ...interface{}) (Result, error) {
	return stmt.SelectOneContext(context.Background(), dst, args...)
}

func (stmt *Stmt) InsertContext(ctx context.Context, args ...interface{}) (Result, error) {
	return insertContext(ctx, nil, stmt, stmt.query, stmt.parseFieldName, stmt.mapping, args...)
}

func (stmt *Stmt) Insert(args ...interface{}) (Result, error) {
	return stmt.InsertContext(context.Background(), args...)
}

func (stmt *Stmt) UpdateContext(ctx context.Context, args ...interface{}) (Result, error) {
	return updateByExecContext(ctx, nil, stmt, stmt.parseFieldName, stmt.mapping, stmt.query, args...)
}

func (stmt *Stmt) Update(args ...interface{}) (Result, error) {
	return stmt.UpdateContext(context.Background(), args...)
}

func (stmt *Stmt) DeleteContext(ctx context.Context, args ...interface{}) (Result, error) {
	result, err := stmt.Stmt.ExecContext(ctx, args...)
	return newResult(result, stmt.query, args), err
}

func (stmt *Stmt) Delete(args ...interface{}) (Result, error) {
	return stmt.DeleteContext(context.Background(), args...)
}

func NewStmt(db *DB, stmt *sql.Stmt, query string) *Stmt {
	return &Stmt{
		DB:    db,
		Stmt:  stmt,
		query: query,
	}
}
