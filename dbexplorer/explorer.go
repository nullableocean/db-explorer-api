package dbexplorer

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"reflect"
	"strconv"
	"strings"
)

type SqlExplorer interface {
	GetTables() ([]string, error)
	GetRecords(table string, offset int, limit int) ([]map[string]interface{}, error)
	GetRecord(table string, id int) (map[string]interface{}, error)
	CreateRecord(table string, data map[string]interface{}) (id int, err error)
	UpdateRecord(table string, id int, data map[string]interface{}) (updated int, err error)
	DeleteRecord(table string, id int) (deleted int, err error)
	HasTable(table string) bool
	ValidateCreateData(table string, data map[string]interface{}) error
	ValidateUpdateData(table string, data map[string]interface{}) error
}

type TableField struct {
	Name       string
	Type       reflect.Kind
	IsNullable bool
	IsPrimary  bool
}

type Explorer struct {
	db          *sql.DB
	tables      map[string]map[string]*TableField
	tableFields map[string][]*TableField
	tableNames  []string
}

func NewSqlExplorer(db *sql.DB) SqlExplorer {
	exp := &Explorer{
		db:          db,
		tables:      make(map[string]map[string]*TableField),
		tableFields: make(map[string][]*TableField),
		tableNames:  make([]string, 0),
	}

	exp.Init()
	return exp
}

func (exp *Explorer) Init() {
	exp.browseTables()

	fmt.Printf("\nexplorer inited...\n\n")
}

func (exp *Explorer) browseTables() {
	rows, err := exp.db.Query("SHOW TABLES")
	if err != nil {
		log.Fatalln(err)
	}
	defer rows.Close()

	for rows.Next() {
		var table string
		rows.Scan(&table)

		exp.tables[table] = make(map[string]*TableField)
		exp.tableNames = append(exp.tableNames, table)

		exp.browseColumns(table)
	}
}

func (exp *Explorer) browseColumns(table string) {
	query := fmt.Sprintf("SHOW FULL COLUMNS FROM `%s`", table)
	rows, err := exp.db.Query(query)
	if err != nil {
		log.Fatalln(err)
	}
	defer rows.Close()

	for rows.Next() {
		col := column{}
		rows.Scan(
			&col.Field,
			&col.Type,
			&col.Collation,
			&col.Null,
			&col.Key,
			&col.Default,
			&col.Extra,
			&col.Privileges,
			&col.Comment,
		)

		exp.saveField(table, &col)
	}
}

func (exp *Explorer) saveField(table string, col *column) {
	field := TableField{
		Name:       col.GetName(),
		Type:       col.GetType(),
		IsNullable: col.IsNullable(),
		IsPrimary:  col.IsPrimary(),
	}

	exp.tables[table][field.Name] = &field
	exp.tableFields[table] = append(exp.tableFields[table], &field)
}

func (exp *Explorer) getField(table string, fieldName string) *TableField {
	return exp.tables[table][fieldName]
}

func (exp *Explorer) getPrimaryKeyField(table string) *TableField {
	for _, v := range exp.tables[table] {
		if v.IsPrimary {
			return v
		}
	}

	return nil
}

func (exp *Explorer) GetTables() ([]string, error) {
	return exp.tableNames, nil
}

func (exp *Explorer) HasTable(table string) bool {
	_, has := exp.tables[table]

	return has
}

func (exp *Explorer) GetRecords(table string, offset int, limit int) ([]map[string]interface{}, error) {
	if !exp.HasTable(table) {
		return nil, ErrTableNotFound
	}

	query := fmt.Sprintf("SELECT * FROM %s LIMIT %d OFFSET %d", table, limit, offset)
	rows, err := exp.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := exp.scanRecords(table, rows)

	return result, nil
}

func (exp *Explorer) GetRecord(table string, id int) (map[string]interface{}, error) {
	if !exp.HasTable(table) {
		return nil, ErrTableNotFound
	}

	primaryField := exp.getPrimaryKeyField(table)
	if primaryField == nil {
		return nil, ErrRecordNotFound
	}

	query := fmt.Sprintf("SELECT * FROM %s WHERE %s = %d", table, primaryField.Name, id)
	rows, err := exp.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := exp.scanRecords(table, rows)
	if len(res) == 0 {
		return nil, ErrRecordNotFound
	}

	return res[0], nil
}

func (exp *Explorer) CreateRecord(table string, data map[string]interface{}) (id int, err error) {
	if !exp.HasTable(table) {
		return id, ErrTableNotFound
	}

	if err := exp.ValidateCreateData(table, data); err != nil {
		return id, err
	}

	primaryField := exp.getPrimaryKeyField(table)

	values := []interface{}{}

	fNames := make([]string, 0, len(exp.tableFields[table]))
	for _, field := range exp.tableFields[table] {
		if field.IsPrimary {
			continue
		}

		val, ex := data[field.Name]
		if !ex {
			values = append(values, "")
		} else {
			values = append(values, val)
		}

		fNames = append(fNames, field.Name)
	}

	fieldsPlaceholders := strings.Join(fNames, ",")
	valuesPlaceholders := strings.TrimSuffix(strings.Repeat("?,", len(values)), ",")

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) RETURNING %s", table, fieldsPlaceholders, valuesPlaceholders, primaryField.Name)
	err = exp.db.QueryRow(query, values...).Scan(&id)

	return id, err
}

func (exp *Explorer) UpdateRecord(table string, id int, data map[string]interface{}) (updated int, err error) {
	if !exp.HasTable(table) {
		return updated, ErrTableNotFound
	}

	if err := exp.ValidateUpdateData(table, data); err != nil {
		return updated, err
	}

	_, err = exp.GetRecord(table, id)
	if err != nil {
		return updated, err
	}

	primaryField := exp.getPrimaryKeyField(table)

	values := []interface{}{}
	placeholderBuilder := strings.Builder{}
	for fname, val := range data {
		field := exp.getField(table, fname)
		if field.IsPrimary {
			continue
		}

		values = append(values, val)
		placeholderBuilder.WriteString(fmt.Sprintf("%s = ?,", fname))
	}
	valuesPlaceholder := strings.TrimSuffix(placeholderBuilder.String(), ",")

	query := fmt.Sprintf("UPDATE %s SET %s WHERE %s = %d", table, valuesPlaceholder, primaryField.Name, id)
	res, err := exp.db.Exec(query, values...)
	if err != nil {
		return updated, nil
	}

	affected, err := res.RowsAffected()
	return int(affected), err
}

func (exp *Explorer) DeleteRecord(table string, id int) (deleted int, err error) {
	if !exp.HasTable(table) {
		return deleted, ErrTableNotFound
	}

	primaryField := exp.getPrimaryKeyField(table)

	query := fmt.Sprintf("DELETE FROM %s WHERE %s = %d", table, primaryField.Name, id)
	res, err := exp.db.Exec(query)
	if err != nil {
		return deleted, nil
	}

	affected, err := res.RowsAffected()
	return int(affected), err
}

func (exp *Explorer) ValidateCreateData(table string, data map[string]interface{}) error {
	for _, field := range exp.tableFields[table] {
		if field.IsPrimary {
			continue
		}

		val, ex := data[field.Name]

		if !field.IsNullable && !ex {
			return errors.New("need required field " + field.Name)
		}

		errMsg := fmt.Sprintf("field %s have invalid type", field.Name)

		if !field.IsNullable && val == nil {
			return errors.New(errMsg)
		} else if val == nil {
			continue
		}

		refValue := reflect.ValueOf(val)
		if field.Type != refValue.Kind() {
			return errors.New(errMsg)
		}
	}

	for fname := range data {
		field := exp.getField(table, fname)
		if field == nil {
			return errors.New("undefined field: " + fname)
		}

		if field.IsPrimary {
			continue
		}
	}

	return nil
}

func (exp *Explorer) ValidateUpdateData(table string, data map[string]interface{}) error {
	for fname, val := range data {
		field := exp.getField(table, fname)
		if field == nil {
			return errors.New("undefined field: " + fname)
		}

		errMsg := fmt.Sprintf("field %s have invalid type", field.Name)

		if field.IsPrimary {
			return errors.New(errMsg)
		}

		if !field.IsNullable && val == nil {
			return errors.New(errMsg)
		} else if val == nil {
			continue
		}

		refValue := reflect.ValueOf(val)
		if field.Type != refValue.Kind() {
			return errors.New(errMsg)
		}
	}

	return nil
}

type recordField struct {
	Name  string
	Value sql.NullString
}

type record struct {
	Fields []*recordField
}

func (exp *Explorer) scanRecords(table string, rows *sql.Rows) []map[string]interface{} {
	cols, _ := rows.Columns()

	record := record{Fields: make([]*recordField, len(cols))}
	addrs := make([]interface{}, len(cols))

	for i, col := range cols {
		record.Fields[i] = &recordField{Name: col}
		addrs[i] = &record.Fields[i].Value
	}

	result := []map[string]interface{}{}
	for rows.Next() {
		rows.Scan(addrs...)

		result = append(result, exp.makeRecordMap(table, record))
	}

	return result
}

func (exp *Explorer) makeRecordMap(table string, rec record) map[string]interface{} {
	recMap := make(map[string]interface{})

	for _, f := range rec.Fields {
		recMap[f.Name] = exp.getRecordFieldValue(table, f)
	}

	return recMap
}

func (exp *Explorer) getRecordFieldValue(table string, f *recordField) interface{} {
	tField := exp.getField(table, f.Name)
	fvalue := f.Value.String
	if tField.IsNullable && fvalue == "" {
		return nil
	}

	var value interface{}

	switch tField.Type {
	case reflect.Int:
		val, err := strconv.Atoi(fvalue)
		if err != nil {
			value = fvalue
		} else {
			value = val
		}
	case reflect.Float64:
		val, err := strconv.ParseFloat(fvalue, 64)
		if err != nil {
			value = fvalue
		} else {
			value = val
		}
	default:
		value = fvalue
	}

	return value
}
