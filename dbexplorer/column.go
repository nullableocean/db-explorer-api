package dbexplorer

import (
	"database/sql"
	"reflect"
	"strings"
)

type column struct {
	Field      string
	Type       string
	Collation  sql.NullString
	Null       sql.NullString
	Key        sql.NullString
	Default    sql.NullString
	Extra      sql.NullString
	Privileges sql.NullString
	Comment    sql.NullString
}

func (col *column) GetName() string {
	return col.Field
}

func (col *column) IsPrimary() bool {
	return col.Key.String == "PRI"
}

func (col *column) IsNullable() bool {
	return col.Null.String == "YES"
}

func (col *column) GetType() reflect.Kind {
	if strings.HasPrefix(col.Type, "int") {
		return reflect.Int
	}

	return reflect.String
}
