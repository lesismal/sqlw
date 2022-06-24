package sqlw

import (
	"context"
	"database/sql"
	"reflect"
	"sync"
)

type DB struct {
	*sql.DB
	tag             string
	mapping         *sync.Map
	fieldNameParser func(field *reflect.StructField) string
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

func (db *DB) Prepare(query string) (*Stmt, error) {
	return db.PrepareContext(context.Background(), query)
}

func (db *DB) PrepareContext(ctx context.Context, query string) (*Stmt, error) {
	stmt, err := db.DB.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}
	return &Stmt{
		DB:   db,
		Stmt: stmt,
	}, nil
}

func (db *DB) QueryRowContext(ctx context.Context, dst interface{}, query string, args ...interface{}) error {
	return queryRowContext(ctx, db.DB, db.parseFieldName, dst, db.mapping, query, args...)
}

func (db *DB) QueryRow(dst interface{}, query string, args ...interface{}) error {
	return db.QueryRowContext(context.Background(), dst, query, args...)
}

func (db *DB) QueryContext(ctx context.Context, dst interface{}, query string, args ...interface{}) error {
	return queryContext(ctx, db.DB, db.parseFieldName, dst, db.mapping, query, args...)
}

func (db *DB) Query(dst interface{}, query string, args ...interface{}) error {
	return db.QueryContext(context.Background(), dst, query, args...)
}

func (db *DB) InsertContext(ctx context.Context, sqlHead string, data interface{}) (sql.Result, error) {
	return insertContext(ctx, db.DB, nil, sqlHead, nil, data, db.parseFieldName, db.mapping)
}

func (db *DB) Insert(sqlHead string, data interface{}) (sql.Result, error) {
	return db.InsertContext(context.Background(), sqlHead, data)
}

func (db *DB) SetFieldParser(f func(field *reflect.StructField) string) {
	db.fieldNameParser = f
}

func (db *DB) Tag() string {
	return db.tag
}

func (db *DB) Mapping() *sync.Map {
	return db.mapping
}

func (db *DB) parseFieldName(field *reflect.StructField) string {
	if db.fieldNameParser != nil {
		return db.fieldNameParser(field)
	}
	return field.Tag.Get(db.tag)
}

func Open(driverName, dataSourceName string, tag string) (*DB, error) {
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return nil, err
	}
	if tag == "" {
		tag = "db"
	}
	return &DB{
		DB:      db,
		tag:     tag,
		mapping: &sync.Map{},
	}, err
}

func Wrap(db *sql.DB, tag string) *DB {
	if tag == "" {
		tag = "db"
	}
	return &DB{
		DB:      db,
		tag:     tag,
		mapping: &sync.Map{},
	}
}
