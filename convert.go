package sqlw

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"sync"
)

type FieldParser func(field *reflect.StructField) string

func sqlMappingKey(opTyp, query string, typ reflect.Type) string {
	return fmt.Sprintf("%v/%v/%v", opTyp, query, typ.String())
}

func rowsToStruct(rows *sql.Rows, dst interface{}, parser FieldParser, mapping *sync.Map, key string, rawScan bool) error {
	dstTyp := reflect.TypeOf(dst)
	// if !isStructPtr(dstTyp) {
	// 	return fmt.Errorf("[sqlw %v] invalid dest type: %v", opTypSelect, dstTyp)
	// }

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
		if rawScan {
			row := make([]interface{}, len(columns))
			dstValue := reflect.Indirect(reflect.ValueOf(dst))
			for i, fieldName := range columns {
				if fieldIdx, ok := fieldIdxMap[fieldName]; ok {
					row[i] = dstValue.Field(fieldIdx).Addr().Interface()
				}
			}
			if err = rows.Scan(row...); err != nil {
				return err
			}
		} else {
			row := newFields(len(columns))
			defer releaseFields(row)

			if err = rows.Scan(row...); err != nil {
				return err
			}

			dstValue := reflect.Indirect(reflect.ValueOf(dst))
			for i, fieldName := range columns {
				if fieldIdx, ok := fieldIdxMap[fieldName]; ok {
					field := row[i].(*Field)
					field.ToValue(dstValue.Field(fieldIdx))
				}
			}
		}
	}

	return nil
}

func rowsToSlice(rows *sql.Rows, dst interface{}, parser FieldParser, mapping *sync.Map, key string, rawScan bool) error {
	dstTyp := reflect.TypeOf(dst)
	if !isStructSlicePtr(dstTyp) {
		return fmt.Errorf("[sqlw %v] invalid dest type: %v", opTypSelect, dstTyp)
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
	var row []interface{}
	if rawScan {
		row = make([]interface{}, len(columns))
	} else {
		row = newFields(len(columns))
		defer releaseFields(row)
	}

	if dstValue.Len() > 0 {
		dstValue.Set(dstValue.Slice(0, 0))
	}
	for rows.Next() {
		dstElemVal := reflect.Indirect(reflect.New(elemTyp))
		if rawScan {
			for i, fieldName := range columns {
				if fieldIdx, ok := fieldIdxMap[fieldName]; ok {
					row[i] = dstElemVal.Field(fieldIdx).Addr().Interface()
				} else {
					row[i] = &Field{}
				}
			}
		}

		if err = rows.Scan(row...); err != nil {
			return err
		}

		if !rawScan {
			for i, fieldName := range columns {
				if fieldIdx, ok := fieldIdxMap[fieldName]; ok {
					field, ok := row[i].(*Field)
					if ok {
						field.ToValue(dstElemVal.Field(fieldIdx))
					} else {

					}
				}
			}
		}

		if isPtrType {
			dstValue.Set(reflect.Append(dstValue, dstElemVal.Addr()))
		} else {
			dstValue.Set(reflect.Append(dstValue, dstElemVal))
		}
	}

	return nil
}

type InsertInfo struct {
	SqlHead      string
	FieldNames   []string
	FieldIndexes map[string]int
}

func parseInsertFields(sqlHead string) ([]string, map[string]struct{}, error) {
	var fieldNames []string
	var fieldNamesMap map[string]struct{}
	if posBegin := strings.Index(sqlHead, "("); posBegin > 1 { // table name and space, at least 2 characters
		fieldNamesMap = map[string]struct{}{}
		posEnd := strings.Index(sqlHead, ")")
		if posEnd < 0 || posEnd < posBegin {
			return nil, nil, fmt.Errorf("[sqlw %v] invalid sql: %v", opTypInsert, sqlHead)
		}
		fieldsStr := sqlHead[posBegin+1 : posEnd]
		fieldNames = strings.Split(fieldsStr, ",")
		if len(fieldNames) == 0 {
			return nil, nil, fmt.Errorf("[sqlw %v] invalid sql: %v", opTypInsert, sqlHead)
		}
		for i, v := range fieldNames {
			s := strings.TrimSpace(v)
			fieldNames[i] = s
			fieldNamesMap[s] = struct{}{}
		}
	}
	return fieldNames, fieldNamesMap, nil
}

func getInsertModelInfo(sqlHead string, dataTyp reflect.Type, mapping *sync.Map, parser FieldParser) (*InsertInfo, error) {
	var err error
	var info *InsertInfo
	var fieldNames []string
	var fieldNamesMap map[string]struct{}
	var key = sqlMappingKey(opTypInsert, sqlHead, dataTyp)
	var stored, ok = mapping.Load(key)
	if ok {
		info = stored.(*InsertInfo)
	} else {
		info = &InsertInfo{
			FieldIndexes: map[string]int{},
		}
		fieldNames, fieldNamesMap, err = parseInsertFields(sqlHead)
		if err != nil {
			return nil, err
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

			if !strings.Contains(sqlHead, opTypInsert) {
				sqlHead = "insert into " + sqlHead
			}

			info.SqlHead = sqlHead
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

	return info, nil
}

func insertContext(ctx context.Context, selector Selector, stmt *Stmt, sqlHead string, parser FieldParser, mapping *sync.Map, args ...interface{}) (Result, error) {
	isStmt := (stmt != nil)
	if !isStmt && sqlHead == "" {
		return nil, fmt.Errorf("[sqlw %v] invalid sql head: %v", opTypInsert, sqlHead)
	}

	sqlHead = strings.ToLower(sqlHead)
	if strings.Contains(sqlHead, opTypSelect) ||
		strings.Contains(sqlHead, opTypUpdate) ||
		strings.Contains(sqlHead, opTypDelete) {
		return nil, fmt.Errorf("[sqlw %v] invalid sql head: %v", opTypInsert, sqlHead)
	}

	var raw = false
	var data interface{}
	var dataTyp reflect.Type
	if len(args) == 1 {
		data = args[0]
		dataTyp = reflect.TypeOf(data)
		if !isInsertable(dataTyp) {
			raw = true
		}
	} else {
		raw = true
	}

	if raw {
		if !isStmt {
			var sqlTail string
			if !strings.Contains(sqlHead, "values") {
				sqlTail = " values"
			}
			if len(args) > 0 {
				fieldNames, _, err := parseInsertFields(sqlHead)
				if err != nil {
					return nil, err
				}
				for i := 0; i < len(args); {
					sqlTail += "("
					for j := 0; j < len(fieldNames); j++ {
						sqlTail += "?"
						if j != len(fieldNames)-1 {
							sqlTail += ","
						}
						i++
					}
					if i != len(args) {
						sqlTail += "),"
					} else {
						sqlTail += ")"
					}
				}
			}
			result, err := selector.ExecContext(ctx, sqlHead+sqlTail, args...)
			return newResult(result, sqlHead, args), err
		}

		result, err := stmt.ExecContext(ctx, args...)
		return newResult(result, stmt.query, args), err
	}

	var err error
	var sqlTail string
	var fieldValues []interface{}
	var insertItems []reflect.Value
	var dataVal = reflect.ValueOf(data)

	// if !isStmt {
	if !strings.Contains(sqlHead, "values") {
		sqlTail = " values"
	}

	info, err := getInsertModelInfo(sqlHead, dataTyp, mapping, parser)
	if err != nil {
		return nil, err
	}
	// }

INIT_FIELD_VALUES:
	switch dataTyp.Kind() {
	case reflect.Struct:
		insertItems = append(insertItems, dataVal)
	case reflect.Ptr:
		dataTyp = dataTyp.Elem()
		dataVal = dataVal.Elem()
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
			insertItems = append(insertItems, val)
		}
	default:
	}

	if !isStmt {
		for i, item := range insertItems {
			sqlTail += "("
			for j, fieldName := range info.FieldNames {
				if idx, ok := info.FieldIndexes[fieldName]; ok {
					fieldValues = append(fieldValues, item.Field(idx).Interface())
					sqlTail += "?"
					if j != len(info.FieldNames)-1 {
						sqlTail += ","
					}
				}
			}
			sqlTail += ")"
			if i != len(insertItems)-1 {
				sqlTail += ","
			}
		}

		query := info.SqlHead + sqlTail
		result, err := selector.ExecContext(ctx, query, fieldValues...)
		return newResult(result, query, fieldValues), err
	}

	for _, item := range insertItems {
		for _, fieldName := range info.FieldNames {
			if idx, ok := info.FieldIndexes[fieldName]; ok {
				fieldValues = append(fieldValues, item.Field(idx).Interface())
			}
		}
	}

	result, err := stmt.ExecContext(ctx, fieldValues...)
	return newResult(result, stmt.query, fieldValues), err
}

func getUpdateModelInfo(sqlHead string, dataTyp reflect.Type, mapping *sync.Map, parser FieldParser) (*InsertInfo, error) {
	var info *InsertInfo
	var fieldNames []string
	var fieldNamesMap map[string]struct{}
	var key = sqlMappingKey(opTypInsert, sqlHead, dataTyp)
	var stored, ok = mapping.Load(key)
	if ok {
		info = stored.(*InsertInfo)
	} else {
		info = &InsertInfo{
			FieldIndexes: map[string]int{},
		}
		if posBegin := strings.Index(sqlHead, "set"); posBegin > 1 { // table name and space, at least 2 characters
			fieldNamesMap = map[string]struct{}{}
			posEnd := strings.Index(sqlHead, "where")
			if posEnd < 0 {
				posEnd = len(sqlHead)
			}
			if posEnd < posBegin {
				return nil, fmt.Errorf("[sqlw %v] invalid sql: %v", opTypUpdate, sqlHead)
			}
			fieldsStr := sqlHead[posBegin+3 : posEnd]
			fields := strings.Split(fieldsStr, ",")
			for _, v := range fields {
				arr := strings.Split(v, "=")
				// if len(arr) == 2 && strings.TrimSpace(arr[1]) == "?" {
				if len(arr) == 2 {
					fieldName := strings.TrimSpace(arr[0])
					fieldNames = append(fieldNames, fieldName)
					fieldNamesMap[fieldName] = struct{}{}
				}
			}
		} else {
			return nil, fmt.Errorf("[sqlw %v] invalid sql: %v", opTypUpdate, sqlHead)
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

				if !strings.Contains(sqlHead, "set") {
					sqlHead = " set " + sqlHead
				}

				for i, fieldName := range fieldNames {
					sqlHead += (fieldName + "=?")
					if i != len(fieldNames)-1 {
						sqlHead += ","
					}
				}
			} else {
				for i := 0; i < typ.NumField(); i++ {
					strField := typ.Field(i)
					fieldName := strings.ToLower(parser(&strField))
					if _, ok := fieldNamesMap[fieldName]; ok {
						info.FieldIndexes[fieldName] = i
					}
				}
			}

			if !strings.Contains(sqlHead, opTypUpdate) {
				sqlHead = opTypUpdate + sqlHead
			}

			info.SqlHead = sqlHead
			info.FieldNames = fieldNames
			mapping.Store(key, info)
		}

		switch dataTyp.Kind() {
		case reflect.Struct:
			initFiedNames(dataTyp)
		case reflect.Ptr:
			dataTyp = dataTyp.Elem()
			initFiedNames(dataTyp)
		default:
		}
	}

	return info, nil
}

func updateContext(ctx context.Context, selector Selector, stmt *Stmt, parser FieldParser, mapping *sync.Map, sqlHead string, data interface{}, args ...interface{}) (Result, error) {
	isStmt := (stmt != nil)
	if !isStmt && sqlHead == "" {
		return nil, fmt.Errorf("[sqlw %v] invalid sql head: %v", opTypInsert, sqlHead)
	}

	sqlHead = strings.ToLower(sqlHead)
	if strings.Contains(sqlHead, opTypSelect) ||
		strings.Contains(sqlHead, opTypInsert) ||
		strings.Contains(sqlHead, opTypDelete) {
		return nil, fmt.Errorf("[sqlw %v] invalid sql head: %v", opTypUpdate, sqlHead)
	}

	var err error
	var fieldValues []interface{}
	var dataTyp = reflect.TypeOf(data)
	var dataVal = reflect.ValueOf(data)

	info, err := getUpdateModelInfo(sqlHead, dataTyp, mapping, parser)
	if err != nil {
		return nil, err
	}

	if isStructPtr(dataTyp) {
		dataVal = dataVal.Elem()
	}

	fieldValues = make([]interface{}, len(info.FieldNames)+len(args))[0:0]
	for _, fieldName := range info.FieldNames {
		if idx, ok := info.FieldIndexes[fieldName]; ok {
			fieldValues = append(fieldValues, dataVal.Field(idx).Interface())
		}
	}

	if len(args) > 0 {
		fieldValues = append(fieldValues, args...)
	}

	if !isStmt {
		query := info.SqlHead
		result, err := selector.ExecContext(ctx, query, fieldValues...)
		return newResult(result, query, fieldValues), err
	}

	result, err := stmt.ExecContext(ctx, fieldValues...)
	return newResult(result, stmt.query, fieldValues), err
}

func isStruct(t reflect.Type) bool {
	return t.Kind() == reflect.Struct
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
