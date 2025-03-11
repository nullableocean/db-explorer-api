package sqlhelp

import (
	"database/sql"
	"reflect"
	"strings"
)

// result is slice of targets
// []<target_type>
func ScanIntoStruct(rows *sql.Rows, target interface{}) (result interface{}, err error) {
	targetVal := reflect.ValueOf(target)
	if targetVal.Kind() == reflect.Ptr {
		targetVal = targetVal.Elem()
	}
	if targetVal.Kind() != reflect.Struct {
		return
	}

	colNames, _ := rows.Columns()
	colTypes, _ := rows.ColumnTypes()

	refs := []interface{}{}
	fieldVal := reflect.Value{}
	var placeholder interface{}

	for i, colName := range colNames {
		colNameParts := strings.Split(colName, ".")
		fieldVal = targetVal.FieldByName(colNameParts[0])

		if fieldVal.IsValid() && fieldVal.Kind() == reflect.Struct && len(colNameParts) > 1 {
			var namePart string
			for _, namePart = range colNameParts[1:] {
				compFunc := matchColName(namePart)
				fieldVal = fieldVal.FieldByNameFunc(compFunc)
			}
		}

		if !fieldVal.IsValid() || !colTypes[i].ScanType().ConvertibleTo(fieldVal.Type()) {
			refs = append(refs, &placeholder)
		}
		if fieldVal.Kind() != reflect.Ptr && fieldVal.CanAddr() {
			fieldVal = fieldVal.Addr()
			refs = append(refs, fieldVal.Interface())
		}
	}

	resultSlice := reflect.MakeSlice(reflect.SliceOf(targetVal.Type()), 0, 10)
	for rows.Next() {
		err = rows.Scan(refs...)
		if err != nil {
			break
		}
		resultSlice = reflect.Append(resultSlice, targetVal)
	}

	result = resultSlice.Interface()

	return result, err
}

func matchColName(colName string) func(string) bool {
	return func(fieldName string) bool {
		return strings.EqualFold(colName, fieldName)
	}
}
