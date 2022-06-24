package sqlw

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"sync"
)

func sqlKey(query string, dst interface{}) string {
	return query + reflect.TypeOf(dst).String()
}

func rowsToStruct(rows *sql.Rows, dst interface{}, parser func(field *reflect.StructField) string, mapping *sync.Map, key string) error {
	dstTyp := reflect.TypeOf(dst)
	if !isStructPtr(dstTyp) {
		return fmt.Errorf("[sqlw query] invalid dest type: %v", dstTyp)
	}

	columns, err := rows.Columns()
	if err != nil {
		return err
	}

	for i, v := range columns {
		columns[i] = strings.ToLower(v)
	}

	var elemTyp = dstTyp.Elem()
	var fieldIdxMap map[string]int
	var stored, ok = mapping.Load(key)
	if ok {
		fieldIdxMap = stored.(map[string]int)
	} else {
		fieldIdxMap = map[string]int{}
		existsMap := map[string]bool{}
		for _, fieldName := range columns {
			existsMap[fieldName] = true
		}
		for j := 0; j < elemTyp.NumField(); j++ {
			strField := elemTyp.Field(j)
			fieldName := strings.ToLower(parser(&strField))
			if existsMap[fieldName] {
				fieldIdxMap[fieldName] = j
			}
		}
		mapping.Store(key, fieldIdxMap)
	}
	if rows.Next() {
		row := newFields(len(columns))
		if err = rows.Scan(row...); err != nil {
			releaseFields(row)
			return err
		}

		dstValue := reflect.Indirect(reflect.ValueOf(dst))
		for i, fieldName := range columns {
			if fieldIdx, ok := fieldIdxMap[fieldName]; ok {
				field := row[i].(*Field)
				field.ToValue(dstValue.Field(fieldIdx))
			}
		}
		releaseFields(row)
	}

	return nil
}

func rowsToSlice(rows *sql.Rows, dst interface{}, parser func(field *reflect.StructField) string, mapping *sync.Map, key string) error {
	dstTyp := reflect.TypeOf(dst)
	if !isStructSlicePtr(dstTyp) {
		return fmt.Errorf("[sqlw query] invalid dest type: %v", dstTyp)
	}

	columns, err := rows.Columns()
	if err != nil {
		return err
	}

	for i, v := range columns {
		columns[i] = strings.ToLower(v)
	}

	elemTyp := dstTyp.Elem().Elem()
	isPtrType := elemTyp.Kind() == reflect.Ptr
	if isPtrType {
		elemTyp = elemTyp.Elem()
	}
	var fieldIdxMap map[string]int
	var stored, ok = mapping.Load(key)
	if ok {
		fieldIdxMap = stored.(map[string]int)
	} else {
		fieldIdxMap = map[string]int{}
		existsMap := map[string]bool{}
		for _, fieldName := range columns {
			existsMap[fieldName] = true
		}
		for j := 0; j < elemTyp.NumField(); j++ {
			strField := elemTyp.Field(j)
			fieldName := strings.ToLower(parser(&strField))
			if existsMap[fieldName] {
				fieldIdxMap[fieldName] = j
			}
		}
		mapping.Store(key, fieldIdxMap)
	}

	dstValue := reflect.Indirect(reflect.ValueOf(dst))
	for rows.Next() {
		row := newFields(len(columns))
		if err = rows.Scan(row...); err != nil {
			releaseFields(row)
			return err
		}

		dstElemVal := reflect.Indirect(reflect.New(elemTyp))
		for i, fieldName := range columns {
			if fieldIdx, ok := fieldIdxMap[fieldName]; ok {
				field := row[i].(*Field)
				field.ToValue(dstElemVal.Field(fieldIdx))
			}
		}

		if isPtrType {
			dstValue.Set(reflect.Append(dstValue, dstElemVal.Addr()))
		} else {
			dstValue.Set(reflect.Append(dstValue, dstElemVal))
		}

		releaseFields(row)
	}

	return nil
}

func insertContext(ctx context.Context, selector Selector, sqlHead string, data interface{}, parser func(field *reflect.StructField) string, mapping *sync.Map) (sql.Result, error) {
	if sqlHead == "" {
		return nil, fmt.Errorf("[sqlw insert] invalid sql: %v", sqlHead)
	}
	if strings.Index(sqlHead, "values") >= 0 {
		return nil, fmt.Errorf("[sqlw insert] invalid sql, should not contains \"values\": %v", sqlHead)
	}
	sqlHead = strings.ToLower(sqlHead)

	dataTyp := reflect.TypeOf(data)
	if !isInsertable(dataTyp) {
		return nil, fmt.Errorf("[sqlw insert] invalid dest type: %v", dataTyp)
	}

	type InsertInfo struct {
		SqlStr       string
		FieldNames   []string
		FieldIndexes map[string]int
	}

	var info *InsertInfo
	var fieldNames []string
	var fieldNamesMap map[string]struct{}
	var fieldValues []interface{}
	var insertItems []reflect.Value
	var sqlTail = " values"
	var key = sqlKey(sqlHead+"insert", data)
	var stored, ok = mapping.Load(key)
	if ok {
		info = stored.(*InsertInfo)
	} else {
		info = &InsertInfo{
			SqlStr:       "",
			FieldIndexes: map[string]int{},
		}
		if posBegin := strings.Index(sqlHead, "("); posBegin > 1 { // table name and space, at least 2 characters
			fieldNamesMap = map[string]struct{}{}
			posEnd := strings.Index(sqlHead, ")")
			if posEnd < 0 || posEnd < posBegin {
				return nil, fmt.Errorf("[sqlw insert] invalid sql: %v", sqlHead)
			}
			fieldsStr := sqlHead[posBegin+1 : posEnd]
			fieldNames = strings.Split(fieldsStr, ",")
			if len(fieldNames) == 0 {
				return nil, fmt.Errorf("[sqlw insert] invalid sql: %v", sqlHead)
			}
			for i, v := range fieldNames {
				s := strings.ToLower(strings.TrimSpace(v))
				fieldNames[i] = s
				fieldNamesMap[s] = struct{}{}
			}

		}
		initFiedNames := func(typ reflect.Type) {
			if len(fieldNames) == 0 {
				for i := 0; i < typ.NumField(); i++ {
					strField := typ.Field(i)
					fieldName := strings.ToLower(parser(&strField))
					if fieldName != "" {
						fieldNames = append(fieldNames, fieldName)
						info.FieldIndexes[fieldName] = i
					}
				}
				sqlHead += "("
				for i, fieldName := range fieldNames {
					sqlHead += fieldName
					if i != len(fieldNames)-1 {
						sqlHead += ","
					}
				}
				sqlHead += ")"
			} else {
				for i := 0; i < typ.NumField(); i++ {
					strField := typ.Field(i)
					fieldName := strings.ToLower(parser(&strField))
					if _, ok := fieldNamesMap[fieldName]; ok {
						info.FieldIndexes[fieldName] = i
					}
				}
			}

			info.SqlStr = sqlHead
			info.FieldNames = fieldNames
			mapping.Store(key, info)
		}
		typ := dataTyp
	INIT_FIELD_NAMES:
		switch typ.Kind() {
		case reflect.Struct:
			initFiedNames(typ)
		case reflect.Ptr:
			typ = typ.Elem()
			goto INIT_FIELD_NAMES
		case reflect.Slice:
			typ = typ.Elem()
			if typ.Kind() == reflect.Ptr {
				typ = typ.Elem()
			}
			initFiedNames(typ)
		default:
		}
	}

	dataVal := reflect.ValueOf(data)

INIT_FIELD_VALUES:
	switch dataTyp.Kind() {
	case reflect.Struct:
		insertItems = append(insertItems, dataVal)
	case reflect.Ptr:
		dataTyp = dataTyp.Elem()
		goto INIT_FIELD_VALUES
	case reflect.Slice:
		isPtrElem := false
		dataTyp = dataTyp.Elem()
		if dataTyp.Kind() == reflect.Ptr {
			isPtrElem = true
		}

		for i := 0; i < dataVal.Len(); i++ {
			val := dataVal.Index(i)
			if isPtrElem {
				val = val.Elem()
			}
			insertItems = append(insertItems, dataVal)
		}
	default:
	}

	for _, item := range insertItems {
		sqlTail += "("
		for i, fieldName := range info.FieldNames {
			if idx, ok := info.FieldIndexes[fieldName]; ok {
				fieldValues = append(fieldValues, item.Field(idx).Interface())
				sqlTail += "?"
				if i != len(info.FieldNames)-1 {
					sqlTail += ","
				}
			}
		}
		sqlTail += ")"
	}

	if !strings.Contains(sqlHead, "insert") {
		sqlHead = "insert into " + sqlHead
	}

	return selector.ExecContext(ctx, sqlHead+sqlTail, fieldValues...)
}

func isStructPtr(t reflect.Type) bool {
	return t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.Struct
}

func isStructSlicePtr(t reflect.Type) bool {
	elem := t.Elem()
	if t.Kind() == reflect.Ptr && elem.Kind() == reflect.Slice {
		sliceElem := elem.Elem()
		if sliceElem.Kind() == reflect.Struct || isStructPtr(sliceElem) {
			return true
		}
	}
	return false
}

func isInsertable(t reflect.Type) bool {
	kind := t.Kind()
	if kind == reflect.Struct || (kind == reflect.Ptr && t.Elem().Kind() == reflect.Struct) {
		return true
	}

	if kind == reflect.Slice {
		kind := t.Elem().Kind()
		if kind == reflect.Struct || (kind == reflect.Ptr && t.Elem().Elem().Kind() == reflect.Struct) {
			return true
		}
	}

	return isStructSlicePtr(t)
}
