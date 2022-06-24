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

func (stmt *Stmt) QueryRowContext(ctx context.Context, dst interface{}, args ...interface{}) error {
	if dst == nil {
		return fmt.Errorf("invalid dest value nil: %v", reflect.TypeOf(dst))
	}

	rows, err := stmt.Stmt.QueryContext(ctx, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	return rowsToStruct(rows, dst, stmt.parseFieldName, stmt.mapping, sqlMappingKey("select", stmt.query, reflect.TypeOf(dst)))
}

func (stmt *Stmt) QueryRow(dst interface{}, args ...interface{}) error {
	return stmt.QueryRowContext(context.Background(), dst, args...)
}

func (stmt *Stmt) QueryContext(ctx context.Context, dst interface{}, args ...interface{}) error {
	rows, err := stmt.Stmt.QueryContext(ctx, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	return rowsToSlice(rows, dst, stmt.parseFieldName, stmt.mapping, sqlMappingKey("select", stmt.query, reflect.TypeOf(dst)))
}

func (stmt *Stmt) Query(dst interface{}, args ...interface{}) error {
	return stmt.QueryContext(context.Background(), dst, args...)
}

func (stmt *Stmt) InsertContext(ctx context.Context, data ...interface{}) (sql.Result, error) {
	return insertContext(ctx, nil, stmt, stmt.query, nil, data, stmt.parseFieldName, stmt.mapping)
}

func (stmt *Stmt) Insert(data ...interface{}) (sql.Result, error) {
	return stmt.InsertContext(context.Background(), data...)
}
