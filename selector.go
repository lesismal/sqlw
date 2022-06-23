package sqlw

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"sync"
)

type Selector interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
}

func queryRowContext(ctx context.Context, selector Selector, parser func(field *reflect.StructField) string, dst interface{}, mapping *sync.Map, query string, args ...interface{}) error {
	if dst == nil {
		return fmt.Errorf("invalid dest value nil: %v", reflect.TypeOf(dst))
	}

	rows, err := selector.QueryContext(ctx, query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	return rowsToStruct(rows, dst, parser, mapping, sqlKey(query, dst))
}

func queryContext(ctx context.Context, selector Selector, parser func(field *reflect.StructField) string, dst interface{}, mapping *sync.Map, query string, args ...interface{}) error {
	rows, err := selector.QueryContext(ctx, query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	return rowsToSlice(rows, dst, parser, mapping, sqlKey(query, dst))
}
