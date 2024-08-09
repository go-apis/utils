package xsheet

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/go-apis/utils/xstring"
)

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

	rnv := reflect.ValueOf(rowNumber)
	mdv := reflect.ValueOf(metadata)

	for i, cell := range cells {
		s := xstring.Clean(cell)
		if len(s) == 0 {
			continue
		}

		if name, ok := m.extra[i]; ok {
			metadata[name] = s
		}

		if prop, ok := m.props[i]; ok {
			e := reflect.ValueOf(s)

			if e.Type().ConvertibleTo(prop.FieldType) {
				v.FieldByName(prop.FieldName).
					Set(e.Convert(prop.FieldType))
			} else {
				// Whoops.
				fmt.Printf("Could not set %s\n", prop.FieldName)
			}
		}

		for _, prop := range m.rowNumberProps {
			v.FieldByName(prop.FieldName).
				Set(rnv.Convert(prop.FieldType))
		}
		for _, prop := range m.extraProps {
			v.FieldByName(prop.FieldName).
				Set(mdv.Convert(prop.FieldType))
		}
	}

	return item, nil
}
