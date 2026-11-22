package csv

import (
	"errors"
)

var ErrNotStruct = errors.New("type must be a struct")