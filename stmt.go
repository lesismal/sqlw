package sqlw

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"sync"
)

type Stmt struct {
	*sql.Stmt
	mapping *sync.Map
	query   string
	parser  func(field *reflect.StructField) string
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

	return rowsToStruct(rows, dst, stmt.parser, stmt.mapping, sqlKey(stmt.query, dst))
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

	return rowsToSlice(rows, dst, stmt.parser, stmt.mapping, sqlKey(stmt.query, dst))
}

func (stmt *Stmt) Query(dst interface{}, args ...interface{}) error {
	return stmt.QueryContext(context.Background(), dst, args...)
}
