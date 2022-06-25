package sqlw

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"sync"
	"time"
)

type Selector interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}

func queryRowContext(ctx context.Context, selector Selector, parser func(field *reflect.StructField) string, dst interface{}, mapping *sync.Map, rawScan bool, query string, args ...interface{}) error {
	if dst == nil {
		return fmt.Errorf("[sqlw %v] invalid dest value nil: %v", opTypSelect, reflect.TypeOf(dst))
	}

	rows, err := selector.QueryContext(ctx, query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	return rowsToStruct(rows, dst, parser, mapping, sqlMappingKey(opTypSelect, query, reflect.TypeOf(dst)), rawScan)
}

func queryContext(ctx context.Context, selector Selector, parser func(field *reflect.StructField) string, dst interface{}, mapping *sync.Map, rawScan bool, query string, args ...interface{}) error {
	rows, err := selector.QueryContext(ctx, query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	if isStructPtr(reflect.TypeOf(dst)) {
		return rowsToStruct(rows, dst, parser, mapping, sqlMappingKey(opTypSelect, query, reflect.TypeOf(dst)), rawScan)
	}

	return rowsToSlice(rows, dst, parser, mapping, sqlMappingKey(opTypSelect, query, reflect.TypeOf(dst)), rawScan)
}

func updateByExecContext(ctx context.Context, selector Selector, stmt *Stmt, parser func(field *reflect.StructField) string, mapping *sync.Map, query string, args ...interface{}) (Result, error) {
	isStructArg := false
	if len(args) == 1 {
		arg := args[0]
		typ := reflect.TypeOf(arg)
		if isStruct(typ) {
			if _, ok := arg.(time.Time); !ok {
				isStructArg = true
			}
		} else if isStructPtr(typ) {
			if _, ok := arg.(*time.Time); !ok {
				isStructArg = true
			}
		}
	}

	if !isStructArg {
		result, err := selector.ExecContext(ctx, query, args...)
		if err != nil {
			return nil, err
		}
		return newResult(result, query), err
	}

	return updateContext(ctx, selector, nil, query, args[0], parser, mapping)
}
