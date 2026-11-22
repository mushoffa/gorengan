package csv

import (
	csvstd "encoding/csv"
	"io"
	"reflect"
)

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
		v := reflect.Indirect(reflect.ValueOf(item))
		for i, c := range cols {
			if row[i], err = formatValue(v.Field(c.index), &cfg); err != nil {
				return err
			}
		}
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	writer.Flush()
	return writer.Error()
}