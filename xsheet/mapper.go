package xsheet

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"

	"github.com/go-apis/utils/xstring"
)

// setStringValue assigns the string s to a struct field, parsing it into the
// field's underlying kind. reflect.Convert does NOT parse strings into numeric
// or boolean types (string is not convertible to int/float/bool), so those
// must go through strconv; otherwise the value is silently dropped.
func setStringValue(field reflect.Value, s string) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(s)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return err
		}
		if field.OverflowInt(n) {
			return fmt.Errorf("value %d overflows %s", n, field.Type())
		}
		field.SetInt(n)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return err
		}
		if field.OverflowUint(n) {
			return fmt.Errorf("value %d overflows %s", n, field.Type())
		}
		field.SetUint(n)
	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return err
		}
		field.SetFloat(f)
	case reflect.Bool:
		b, err := strconv.ParseBool(s)
		if err != nil {
			return err
		}
		field.SetBool(b)
	default:
		// Named string types and the like remain convertible directly.
		e := reflect.ValueOf(s)
		if !e.Type().ConvertibleTo(field.Type()) {
			return fmt.Errorf("cannot assign string to field of type %s", field.Type())
		}
		field.Set(e.Convert(field.Type()))
	}
	return nil
}

type Mapper[T any] interface {
	Parse(rowNumber int, cells []string) (*T, error)
}

type mapper[T any] struct {
	props          map[int]*PropField
	rowNumberProps []*PropField
	extraProps     []*PropField
	extra          map[int]string
}

func (m *mapper[T]) Parse(rowNumber int, cells []string) (item *T, err error) {
	defer func() {
		if r := recover(); r != nil {
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				// Fallback err (per specs, error strings should be lowercase w/o punctuation
				err = errors.New("unknown panic")
			}
		}
	}()

	item = new(T)

	v := reflect.ValueOf(item)
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	metadata := map[string]string{}

	for i, cell := range cells {
		s := xstring.Clean(cell)
		if len(s) == 0 {
			continue
		}

		if name, ok := m.extra[i]; ok {
			metadata[name] = s
		}

		if prop, ok := m.props[i]; ok {
			if err := setStringValue(v.FieldByName(prop.FieldName), s); err != nil {
				return nil, fmt.Errorf("could not set %s: %w", prop.FieldName, err)
			}
		}
	}

	// Row number and extra metadata are per-row, not per-cell: set them once
	// after scanning the cells so they're populated even when leading cells
	// (or the whole row) are empty.
	rnv := reflect.ValueOf(rowNumber)
	for _, prop := range m.rowNumberProps {
		v.FieldByName(prop.FieldName).Set(rnv.Convert(prop.FieldType))
	}
	mdv := reflect.ValueOf(metadata)
	for _, prop := range m.extraProps {
		v.FieldByName(prop.FieldName).Set(mdv.Convert(prop.FieldType))
	}

	return item, nil
}
