package sqlw

import (
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
		return fmt.Errorf("invalid dest type: %v", dstTyp)
	}

	columns, err := rows.Columns()
	if err != nil {
		return err
	}

	for i, v := range columns {
		columns[i] = strings.ToLower(v)
	}

	fmt.Println("columns:", columns)

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
		// fmt.Println("existsMap:", existsMap)
		// for _, fieldName := range columns {
		for j := 0; j < elemTyp.NumField(); j++ {
			strField := elemTyp.Field(j)
			fieldName := strings.ToLower(parser(&strField))
			if existsMap[fieldName] {
				fieldIdxMap[fieldName] = j
				// break
			}
		}
		// fmt.Println("fieldIdxMap:", fieldIdxMap)
		// }

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
		return fmt.Errorf("invalid dest type: %v", dstTyp)
	}

	columns, err := rows.Columns()
	if err != nil {
		return err
	}

	for i, v := range columns {
		columns[i] = strings.ToLower(v)
	}

	fmt.Println("columns:", columns)

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
		// fmt.Println("existsMap:", existsMap)
		// for _, fieldName := range columns {
		for j := 0; j < elemTyp.NumField(); j++ {
			strField := elemTyp.Field(j)
			fieldName := strings.ToLower(parser(&strField))
			if existsMap[fieldName] {
				fieldIdxMap[fieldName] = j
			}
		}
		// fmt.Println("fieldIdxMap:", fieldIdxMap)
		// }

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
