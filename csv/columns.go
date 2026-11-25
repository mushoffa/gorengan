package csv

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
)

// column maps one CSV column to a (possibly nested) struct field.
type column struct {
	index []int  // path through nested structs, used by fieldByIndex
	name  string // header text
}

// columnCache avoids re-running reflection for the same type.
var columnCache sync.Map // map[reflect.Type][]column

// columnsOf returns the CSV columns of struct type t in declaration order.
// Pointer types are unwrapped, and embedded structs are flattened.
// The returned slice is shared and must be treated as read-only.
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

	var cols []column
	collectColumns(t, nil, &cols)

	seen := make(map[string]struct{}, len(cols))
	for _, c := range cols {
		if _, dup := seen[c.name]; dup {
			return nil, fmt.Errorf("%w: %q in %s", ErrDuplicateColumn, c.name, t)
		}
		seen[c.name] = struct{}{}
	}

	columnCache.Store(t, cols)
	return cols, nil
}

// collectColumns walks the fields of t, appending columns to cols.
// prefix is the index path of the struct being walked.
func collectColumns(t reflect.Type, prefix []int, cols *[]column) {
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)

		tag := f.Tag.Get("csv")
		tagName, _, _ := strings.Cut(tag, ",") // room for future options
		if tagName == "-" {
			continue
		}

		// Copy so sibling fields never share a backing array.
		path := make([]int, len(prefix)+1)
		copy(path, prefix)
		path[len(prefix)] = i

		// Flatten embedded structs unless given an explicit name.
		if f.Anonymous && tagName == "" {
			ft := f.Type
			if ft.Kind() == reflect.Pointer {
				ft = ft.Elem()
			}
			if ft.Kind() == reflect.Struct && ft != timeType {
				collectColumns(ft, path, cols)
				continue
			}
		}

		if !f.IsExported() {
			continue
		}

		name := f.Name
		if tagName != "" {
			name = tagName
		}
		*cols = append(*cols, column{index: path, name: name})
	}
}

// headerOf extracts the header row from columns.
func headerOf(cols []column) []string {
	h := make([]string, len(cols))
	for i, c := range cols {
		h[i] = c.name
	}
	return h
}

// fieldByIndex follows an index path through v, dereferencing pointers on
// the way. It returns false if a nil pointer is met (for example a nil
// embedded *Product, or a nil *Person element).
func fieldByIndex(v reflect.Value, index []int) (reflect.Value, bool) {
	for _, i := range index {
		if v.Kind() == reflect.Pointer {
			if v.IsNil() {
				return reflect.Value{}, false
			}
			v = v.Elem()
		}
		v = v.Field(i)
	}
	return v, true
}