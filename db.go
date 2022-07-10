// Copyright 2022 lesismal. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package sqlw

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"sync"
)

type DB struct {
	*sql.DB
	tag                string
	placeholder        string
	placeholderBuilder func(int) string
	rawScan            bool
	mapping            *sync.Map
	fieldNameParser    FieldParser

	ctx     context.Context
	cancel  func()
	isMysql bool
}

func (db *DB) Begin() (*Tx, error) {
	tx, err := db.DB.Begin()
	if err != nil {
		return nil, err
	}
	return &Tx{
		DB: db,
		Tx: tx,
	}, nil
}

func (db *DB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*Tx, error) {
	tx, err := db.DB.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &Tx{
		DB: db,
		Tx: tx,
	}, nil
}

func (db *DB) ExecContext(ctx context.Context, query string, args ...interface{}) (Result, error) {
	result, err := db.DB.ExecContext(ctx, query, args...)
	return newResult(result, query, args), err
}

func (db *DB) Exec(query string, args ...interface{}) (Result, error) {
	return db.ExecContext(db.ctx, query, args...)
}

func (db *DB) Prepare(query string) (*Stmt, error) {
	return db.PrepareContext(db.ctx, query)
}

func (db *DB) PrepareContext(ctx context.Context, query string) (*Stmt, error) {
	stmt, err := db.DB.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}
	return NewStmt(db, stmt, query), nil
}

func (db *DB) QueryRowContext(ctx context.Context, dst interface{}, query string, args ...interface{}) (Result, error) {
	return queryRowContext(ctx, db.DB, db.parseFieldName, dst, db.mapping, db.rawScan, query, args...)
}

func (db *DB) QueryRow(dst interface{}, query string, args ...interface{}) (Result, error) {
	return db.QueryRowContext(db.ctx, dst, query, args...)
}

func (db *DB) QueryContext(ctx context.Context, dst interface{}, query string, args ...interface{}) (Result, error) {
	return queryContext(ctx, db.DB, db.parseFieldName, dst, db.mapping, db.rawScan, query, args...)
}

func (db *DB) Query(dst interface{}, query string, args ...interface{}) (Result, error) {
	return db.QueryContext(db.ctx, dst, query, args...)
}

func (db *DB) SelectContext(ctx context.Context, dst interface{}, query string, args ...interface{}) (Result, error) {
	return db.QueryContext(ctx, dst, query, args...)
}

func (db *DB) Select(dst interface{}, query string, args ...interface{}) (Result, error) {
	return db.QueryContext(db.ctx, dst, query, args...)
}

// deprecated.
// func (db *DB) SelectOneContext(ctx context.Context, dst interface{}, query string, args ...interface{}) (Result, error) {
// 	typ := reflect.TypeOf(dst)
// 	if !isStructPtr(typ) {
// 		return newResult(nil, query, args), fmt.Errorf("[sqlw %v] invalid dest type: %v", opTypSelect, typ)
// 	}
// 	return db.SelectContext(db.ctx, dst, query, args...)
// }

// deprecated.
// func (db *DB) SelectOne(dst interface{}, query string, args ...interface{}) (Result, error) {
// 	return db.SelectOneContext(db.ctx, dst, query, args...)
// }

func (db *DB) InsertContext(ctx context.Context, sqlHead string, args ...interface{}) (Result, error) {
	return insertContext(ctx, db.DB, nil, sqlHead, db, args...)
}

func (db *DB) Insert(sqlHead string, args ...interface{}) (Result, error) {
	return db.InsertContext(db.ctx, sqlHead, args...)
}

func (db *DB) UpdateContext(ctx context.Context, sqlHead string, args ...interface{}) (Result, error) {
	return updateByExecContext(ctx, db.DB, db, nil, sqlHead, args...)
}

func (db *DB) Update(sqlHead string, args ...interface{}) (Result, error) {
	return db.UpdateContext(db.ctx, sqlHead, args...)
}

func (db *DB) DeleteContext(ctx context.Context, query string, args ...interface{}) (Result, error) {
	result, err := db.DB.ExecContext(ctx, query, args...)
	return newResult(result, query, args), err
}

func (db *DB) Delete(query string, args ...interface{}) (Result, error) {
	return db.DeleteContext(db.ctx, query, args...)
}

func (db *DB) SetFieldParser(f FieldParser) {
	db.fieldNameParser = f
}

func (db *DB) Close() error {
	err := db.DB.Close()
	db.cancel()
	return err
}

func (db *DB) Tag() string {
	return db.tag
}

func (db *DB) SetTag(tag string) {
	db.tag = tag
}

func (db *DB) Placeholder() string {
	return db.placeholder
}

func (db *DB) SetPlaceholder(placeholder string) {
	db.placeholder = placeholder
}

func (db *DB) PlaceholderBuilder() func(int) string {
	return db.placeholderBuilder
}

func (db *DB) SetPlaceholderBuilder(placeholderBuilder func(int) string) {
	db.placeholderBuilder = placeholderBuilder
}

func (db *DB) Context() context.Context {
	return db.ctx
}

func (db *DB) SetContext(ctx context.Context) {
	if ctx != nil {
		db.ctx = ctx
		db.cancel = func() {}
	}
}

func (db *DB) Mapping() *sync.Map {
	return db.mapping
}

func (db *DB) RawScan() bool {
	return db.rawScan
}

func (db *DB) SetRawScan(rawScan bool) {
	db.rawScan = rawScan
}

func (db *DB) parseFieldName(field *reflect.StructField) string {
	if db.fieldNameParser != nil {
		return db.fieldNameParser(field)
	}
	return field.Tag.Get(db.tag)
}

func Open(driverName, dataSourceName string, tag string) (*DB, error) {
	return OpenContext(nil, driverName, dataSourceName, tag)
}

func OpenContext(ctx context.Context, driverName, dataSourceName string, tag string) (*DB, error) {
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return nil, err
	}
	return WrapContext(ctx, db, driverName, tag), err
}

func Wrap(db *sql.DB, driverName, tag string) *DB {
	return WrapContext(nil, db, driverName, tag)
}

func WrapContext(ctx context.Context, db *sql.DB, driverName, tag string) *DB {
	sqlwDB := &DB{
		DB:                 db,
		tag:                tag,
		placeholder:        "$",
		placeholderBuilder: func(int) string { return "?" },
		rawScan:            true,
		mapping:            &sync.Map{},
		ctx:                ctx,
		cancel:             func() {},
		isMysql:            true,
	}
	if !strings.Contains(driverName, "mysql") {
		sqlwDB.isMysql = false
		sqlwDB.placeholder = "$"
		sqlwDB.placeholderBuilder = func(i int) string {
			return fmt.Sprintf("$%d", i)
		}
	}
	if ctx == nil {
		sqlwDB.ctx, sqlwDB.cancel = context.WithCancel(context.Background())
	}
	return sqlwDB
}
