package xsheet

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"unicode"

	"github.com/shakinm/xlsReader/xls"
	"github.com/xuri/excelize/v2"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
)

var ErrInvalidFormat = fmt.Errorf("Invalid file format")
var ErrNoSheetsFound = fmt.Errorf("No sheets found")
var ErrNoRowsFound = fmt.Errorf("No rows found")

type Parser[T any] interface {
	Parse(ctx context.Context, fileName string, fileType string, reader io.Reader) ([]*T, error)
}

type parser[T any] struct {
	props Props[T]
}

func (p *parser[T]) getRowValues(s xls.Sheet, row int) ([]string, error) {
	rows, err := s.GetRow(row)
	if err != nil {
		return nil, err
	}

	out := []string{}
	for _, c := range rows.GetCols() {
		out = append(out, c.GetString())
	}
	return out, nil
}

func (p *parser[T]) parseCsv(ctx context.Context, reader io.Reader) ([]*T, error) {
	grr := func(r rune) bool {
		if r == '\n' {
			return false
		}
		return !unicode.IsPrint(r)
	}
	r := transform.NewReader(reader, runes.Remove(runes.Predicate(grr)))

	f := csv.NewReader(r)

	records, err := f.ReadAll()
	if err != nil {
		return nil, err
	}
	rowCount := len(records)
	if rowCount <= 1 {
		return nil, ErrNoRowsFound
	}

	mapper, err := p.props.Headers(records[0])
	if err != nil {
		return nil, err
	}
	items := make([]*T, rowCount-1)
	for i := range items {
		rowNumber := i + 1
		rowValues := records[rowNumber]
		item, err := mapper.Parse(i+1, rowValues)
		if err != nil {
			return nil, fmt.Errorf("Failed to parse row %d: %w", i+1, err)
		}
		items[i] = item
	}
	return items, nil
}
func (p *parser[T]) parseExcelOld(ctx context.Context, reader io.Reader) ([]*T, error) {
	seeker, ok := reader.(io.ReadSeeker)
	if !ok {
		return nil, fmt.Errorf("Expected a ReadSeeker")
	}
	f, err := xls.OpenReader(seeker)
	if err != nil {
		return nil, err
	}
	sheets := f.GetSheets()
	if len(sheets) == 0 {
		return nil, ErrNoSheetsFound
	}
	s := sheets[0]

	rowCount := s.GetNumberRows()
	if rowCount <= 1 {
		return nil, ErrNoRowsFound
	}

	headerValues, err := p.getRowValues(s, 0)
	if err != nil {
		return nil, err
	}

	mapper, err := p.props.Headers(headerValues)
	if err != nil {
		return nil, err
	}

	items := make([]*T, rowCount-1)
	for i := range items {
		rowNumber := i + 1
		rowValues, err := p.getRowValues(s, rowNumber)
		if err != nil {
			return nil, err
		}
		item, err := mapper.Parse(i+1, rowValues)
		if err != nil {
			return nil, fmt.Errorf("Failed to parse row %d: %w", i+1, err)
		}
		items[i] = item
	}

	return items, nil
}
func (p *parser[T]) parseExcel(ctx context.Context, reader io.Reader) ([]*T, error) {
	f, err := excelize.OpenReader(reader)
	if err != nil {
		return nil, err
	}
	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, ErrNoSheetsFound
	}
	// get the first sheet
	sheet := sheets[0]

	// get the rows
	rows, err := f.GetRows(sheet)
	if err != nil {
		return nil, err
	}
	rowCount := len(rows)
	if rowCount <= 1 {
		return nil, ErrNoRowsFound
	}

	mapper, err := p.props.Headers(rows[0])
	if err != nil {
		return nil, err
	}

	items := make([]*T, rowCount-1)
	for i := range items {
		rowNumber := i + 1
		rowValues := rows[rowNumber]
		item, err := mapper.Parse(rowNumber, rowValues)
		if err != nil {
			return nil, fmt.Errorf("Failed to parse row %d: %w", i+1, err)
		}
		items[i] = item
	}
	return items, nil
}

func (p *parser[T]) Parse(ctx context.Context, fileName string, fileType string, reader io.Reader) ([]*T, error) {
	switch {
	case IsCsvFilename(fileName):
		return p.parseCsv(ctx, reader)
	case IsExcelOldFileType(fileType):
		return p.parseExcelOld(ctx, reader)
	case IsExcelFileType(fileType):
		return p.parseExcel(ctx, reader)
	default:
		return nil, ErrInvalidFormat
	}
}

func NewParser[T any]() (Parser[T], error) {
	// build mapper from T.
	props, err := NewProps[T]()
	if err != nil {
		return nil, err
	}

	return &parser[T]{
		props,
	}, nil
}
