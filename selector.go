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

func queryRowContext(ctx context.Context, selector Selector, parser FieldParser, dst interface{}, mapping *sync.Map, rawScan bool, query string, args ...interface{}) error {
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

func queryContext(ctx context.Context, selector Selector, parser FieldParser, dst interface{}, mapping *sync.Map, rawScan bool, query string, args ...interface{}) error {
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

func updateByExecContext(ctx context.Context, selector Selector, stmt *Stmt, parser FieldParser, mapping *sync.Map, query string, args ...interface{}) (Result, error) {
	var obj interface{}
	if len(args) > 0 {
		typ := reflect.TypeOf(args[0])
		if isStruct(typ) {
			if _, ok := args[0].(time.Time); !ok {
				obj = args[0]
				args = args[1:]
			}
		} else if isStructPtr(typ) {
			if _, ok := args[0].(*time.Time); !ok {
				obj = args[0]
				args = args[1:]
			}
		}
	}

	if obj == nil {
		if selector != nil {
			result, err := selector.ExecContext(ctx, query, args...)
			if err != nil {
				return nil, err
			}
			return newResult(result, query, args), err
		}

		result, err := stmt.ExecContext(ctx, args...)
		if err != nil {
			return nil, err
		}
		return newResult(result, query, args), err
	}

	return updateContext(ctx, selector, stmt, parser, mapping, query, obj, args...)
}
