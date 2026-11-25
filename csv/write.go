package csv

import (
	"bytes"
	csvstd "encoding/csv"
	"io"
	"os"
	"path/filepath"
	"reflect"
)

// Write writes data as CSV to w. T must be a struct or a pointer to a
// struct. The header comes from the struct's exported fields; use the `csv`
// tag to rename a field or `csv:"-"` to skip it. Embedded structs are
// flattened.
func Write[T any](w io.Writer, data []T, opts ...Option) error {
	cfg := applyOptions(opts)

	cols, err := columnsOf(reflect.TypeOf((*T)(nil)).Elem())
	if err != nil {
		return err
	}

	writer := csvstd.NewWriter(w)
	writer.Comma = cfg.delimiter
	writer.UseCRLF = cfg.crlf

	if !cfg.noHeader {
		if err := writer.Write(headerOf(cols)); err != nil {
			return err
		}
	}

	row := make([]string, len(cols))
	for _, item := range data {
		v := reflect.ValueOf(item)

		for i, c := range cols {
			fv, ok := fieldByIndex(v, c.index)
			if !ok { // nil item or nil embedded pointer
				row[i] = cfg.nilValue
				continue
			}
			s, err := formatValue(fv, &cfg)
			if err != nil {
				return err
			}
			row[i] = s
		}

		if err := writer.Write(row); err != nil {
			return err
		}
	}

	writer.Flush()
	return writer.Error()
}

// WriteFile writes data to the named file, creating or truncating it.
// Rows are streamed to disk, so it is safe for large slices.
func WriteFile[T any](path, filename string, data []T, opts ...Option) (err error) {
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return err
	}
	f, err := os.Create(filepath.Join(path, filename))
	if err != nil {
		return err
	}
	defer func() {
		if cerr := f.Close(); err == nil {
			err = cerr
		}
	}()

	return Write(f, data, opts...)
}

// Marshal returns the CSV encoding of data as a byte slice.
func Marshal[T any](data []T, opts ...Option) ([]byte, error) {
	var buf bytes.Buffer
	if err := Write(&buf, data, opts...); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// MarshalString returns the CSV encoding of data as a string.
func MarshalString[T any](data []T, opts ...Option) (string, error) {
	b, err := Marshal(data, opts...)
	return string(b), err
}