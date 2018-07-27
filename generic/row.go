package generic

import (
	"errors"
	"fmt"
)

// Token stores the starting and ending positions of a column value (piece of text).
// Positions are relative to the start of the underlying ReadSeeker. Start is inclusive,
// while End is exclusive.
type Token struct {
	Start int
	End   int
}

// ValueSize returns the byte size of a value. Returns an error if value size cannot be calculated.
func (t Token) ValueSize() (int, error) {
	if t.Start == -1 {
		return 0, errors.New("invalid token: Start == -1")
	} else if t.End == -1 {
		return 0, errors.New("invalid token: End == -1")
	} else if t.Start > t.End {
		return 0, fmt.Errorf("invalid token: Start (%d) > End (%d) position", t.Start, t.End)
	}
	return t.End - t.Start, nil
}

// Row stores the starting and ending positions of column values found on this row as token objects.
type Row struct {
	tokens []Token
}

// newRow creates and initializes an empty Row structure
func newRow() *Row {
	r := &Row{}
	return r
}

// ColCount returns the number of columns found in this row.
func (r *Row) ColCount() int {
	return len(r.tokens)
}

// HasColumn returns true if a token for the specified column is available.
func (r *Row) HasColumn(col int) bool {
	return col >= 0 && (col < len(r.tokens))
}

// ValueSize returns the byte size of a value by it's column index. Returns an error if there
// is no token for the given column or if value size cannot be calculated.
func (r *Row) ValueSize(col int) (int, error) {
	if !r.HasColumn(col) {
		return 0, fmt.Errorf("invalid column index %d", col)
	}
	t := r.tokens[col]
	return t.ValueSize()
}

func (r *Row) addToken(start, end int) {
	r.tokens = append(r.tokens, Token{start, end})
}

// Token returns the token for the column with the specified index. An error is returned if
// token for that column is not available.
func (r *Row) Token(col int) (*Token, error) {
	if !r.HasColumn(col) {
		return nil, fmt.Errorf("invalid column index %d", col)
	}
	return &r.tokens[col], nil
}
