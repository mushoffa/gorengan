package csv

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
)

// column maps a CSV column to a struct field.
type column struct {
	index int    // field index in the struct
	name  string // header text
}

// columnCache avoids re-running reflection for the same type.
var columnCache sync.Map // map[reflect.Type][]column

// columnsOf returns the exported, non-skipped fields of struct type t,
// in declaration order. Pointer types are unwrapped.
func columnsOf(t reflect.Type) ([]column, error) {
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("%w: got %s", ErrNotStruct, t.Kind())
	}

	if cached, ok := columnCache.Load(t); ok {
		return cached.([]column), nil
	}

	cols := make([]column, 0, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if !f.IsExported() {
			continue
		}

		name := f.Name
		if tag, ok := f.Tag.Lookup("csv"); ok {
			tagName, _, _ := strings.Cut(tag, ",") // room for future options
			if tagName == "-" {
				continue
			}
			if tagName != "" {
				name = tagName
			}
		}
		cols = append(cols, column{index: i, name: name})
	}

	columnCache.Store(t, cols)
	return cols, nil
}

// headerOf extracts the header row from columns.
func headerOf(cols []column) []string {
	h := make([]string, len(cols))
	for i, c := range cols {
		h[i] = c.name
	}
	return h
}