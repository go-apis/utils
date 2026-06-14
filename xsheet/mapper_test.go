package xsheet

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type typedRow struct {
	Name   string  `sheet:"name"`
	Age    int     `sheet:"age"`
	Score  float64 `sheet:"score"`
	Active bool    `sheet:"active"`
	Row    int     `sheet:";rownumber"`
}

func newTypedMapper(t *testing.T) Mapper[typedRow] {
	t.Helper()
	p, err := NewProps[typedRow]()
	require.NoError(t, err)
	m, err := p.Headers([]string{"name", "age", "score", "active"})
	require.NoError(t, err)
	return m
}

// Numeric and bool columns must actually be parsed, not silently dropped.
// reflect.Convert does not turn "30" into int(30), so this regresses if the
// strconv path is removed.
func TestParse_NumericAndBoolFields(t *testing.T) {
	m := newTypedMapper(t)

	item, err := m.Parse(5, []string{"Alice", "30", "9.5", "true"})
	require.NoError(t, err)

	assert.Equal(t, "Alice", item.Name)
	assert.Equal(t, 30, item.Age)
	assert.Equal(t, 9.5, item.Score)
	assert.True(t, item.Active)
	assert.Equal(t, 5, item.Row)
}

// A non-numeric value in a numeric column should surface an error rather than
// being silently swallowed.
func TestParse_BadNumberReturnsError(t *testing.T) {
	m := newTypedMapper(t)

	_, err := m.Parse(1, []string{"Alice", "not-a-number", "9.5", "true"})
	require.Error(t, err)
}

// The row number is per-row, so it must be set even when every cell is empty.
func TestParse_RowNumberSetOnEmptyRow(t *testing.T) {
	m := newTypedMapper(t)

	item, err := m.Parse(7, []string{"", "", "", ""})
	require.NoError(t, err)
	assert.Equal(t, 7, item.Row)
}
