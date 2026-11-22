package csv

import (
	"encoding"
	"fmt"
	"reflect"
	"strconv"
	"time"
)

var timeType = reflect.TypeOf(time.Time{})

func formatValue(v reflect.Value, cfg *config) (string, error) {
	// nil pointers and interfaces become the configured empty value
	if k := v.Kind(); k == reflect.Pointer || k == reflect.Interface {
		if v.IsNil() {
			return cfg.nilValue, nil
		}
		return formatValue(v.Elem(), cfg)
	}

	// time.Time first: it also implements TextMarshaler, but we want a custom layout
	if v.Type() == timeType {
		return v.Interface().(time.Time).Format(cfg.timeFormat), nil
	}

	// types that know how to render themselves
	if v.CanInterface() {
		if m, ok := v.Interface().(encoding.TextMarshaler); ok {
			b, err := m.MarshalText()
			return string(b), err
		}
	}

	switch v.Kind() {
	case reflect.String:
		return v.String(), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(v.Uint(), 10), nil
	case reflect.Float32:
		return strconv.FormatFloat(v.Float(), 'f', -1, 32), nil
	case reflect.Float64:
		return strconv.FormatFloat(v.Float(), 'f', -1, 64), nil
	case reflect.Bool:
		return strconv.FormatBool(v.Bool()), nil
	default:
		return fmt.Sprint(v.Interface()), nil // Stringers, slices, etc.
	}
}
