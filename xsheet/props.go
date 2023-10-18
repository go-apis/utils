package xsheet

import (
	"fmt"
	"reflect"
	"strings"
)

var ErrPropertyExists = fmt.Errorf("Property exists")

type PropField struct {
	FieldType reflect.Type
	FieldName string
	Name      string
	RowNumber bool
	Aliases   []string
	Extra     bool
	Skip      bool
}

func NewPropField(f reflect.StructField) *PropField {
	aliases := []string{}
	name := f.Name
	extra := false
	skip := false
	rownumber := false

	tag := f.Tag.Get("sheet")
	splits := strings.Split(tag, ";")
	for _, split := range splits {
		inner := strings.Split(split, ":")

		switch {
		case len(inner) == 0:
			continue

		case len(inner) == 1 && inner[0] == "-":
			skip = true
			continue

		case len(inner) == 1 && strings.EqualFold(inner[0], "extra"):
			extra = true
			continue

		case len(inner) == 1 && strings.EqualFold(inner[0], "rownumber"):
			rownumber = true
			continue

		case len(inner) == 1:
			name = strings.ToLower(inner[0])
			continue

		case len(inner) == 2 && strings.EqualFold(inner[0], "alias"):
			aliases = strings.Split(strings.ToLower(inner[1]), ",")
			continue
		}
	}

	return &PropField{
		FieldName: f.Name,
		FieldType: f.Type,
		Name:      name,
		RowNumber: rownumber,
		Aliases:   aliases,
		Extra:     extra,
		Skip:      skip,
	}
}

type Props[T any] interface {
	Headers(headers []string) (Mapper[T], error)
}

type props[T any] struct {
	rowNumbers []*PropField
	extras     []*PropField
	propFields map[string]*PropField
}

func (p *props[T]) Headers(headers []string) (Mapper[T], error) {
	props := make(map[int]*PropField)
	extra := make(map[int]string)

	for i, name := range headers {
		key := strings.ToLower(name)
		if prop, ok := p.propFields[key]; ok {
			props[i] = prop
		} else {
			extra[i] = name
		}
	}

	return &mapper[T]{
		rowNumberProps: p.rowNumbers,
		extraProps:     p.extras,
		props:          props,
		extra:          extra,
	}, nil
}

func NewProps[T any]() (Props[T], error) {
	t := ToType[T]()

	rowNumbers := []*PropField{}
	extras := []*PropField{}
	propFields := map[string]*PropField{}

	fieldCount := t.NumField()
	for i := 0; i < fieldCount; i++ {
		field := t.Field(i)
		prop := NewPropField(field)
		if prop.Skip {
			continue
		}

		if prop.Extra {
			extras = append(extras, prop)
		}
		if prop.RowNumber {
			rowNumbers = append(rowNumbers, prop)
		}

		if len(prop.Name) == 0 {
			continue
		}
		if _, exists := propFields[prop.Name]; exists {
			return nil, fmt.Errorf("Propery %s %w", prop.Name, ErrPropertyExists)
		}
		propFields[prop.Name] = prop

		for _, alias := range prop.Aliases {
			if alias == prop.Name {
				continue
			}
			if _, exists := propFields[alias]; exists {
				return nil, fmt.Errorf("Propery %s %w", alias, ErrPropertyExists)
			}
			propFields[alias] = prop
		}
	}

	return &props[T]{
		rowNumbers: rowNumbers,
		extras:     extras,
		propFields: propFields,
	}, nil
}
