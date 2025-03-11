package dbexplorer

import "errors"

var (
	ErrTableNotFound  = errors.New("unknown table")
	ErrRecordNotFound = errors.New("record not found")
)
