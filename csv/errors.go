package csv

import (
	"errors"
)

var ErrNotStruct = errors.New("csv: type must be a struct")
var ErrDuplicateColumn = errors.New("csv: duplicate column name")
