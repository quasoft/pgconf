package hba

import (
	"fmt"
	"io"
	"io/ioutil"

	"github.com/quasoft/pgconf/generic"
)

// Constants for column indexes
const (
	ConnType = iota
	Database
	User
	Address
	Method
)

// newParams creates param structure with defaults suitable for parsing of pg_hba.conf files:
//  - Whitespace:             space, tab and carriage return
//  - DefaultDelim: 		  tab
//  - Quotes:                 " and '
//  - BackslashEscapedQuotes: true (allows use of \" and \' for escaping of quote characters in values)
//  - DefaultQuote:           "
//  - InlineComment:          #
//  - AlwaysQuoteStrings:	  false
func newParams() generic.Params {
	return generic.Params{
		Whitespace:             " \t\r",
		DefaultDelim:           "\t",
		Quotes:                 `"'`,
		BackslashEscapedQuotes: true,
		DefaultQuote:           '"',
		InlineComment:          '#',
		AlwaysQuoteStrings:     false,
	}
}

// Conf represents configuration file for host-based authentication of PostgreSQL (pg_hba.conf).
type Conf struct {
	*generic.Conf
}

// New creates a new structure for reading/writing to pg_hba.conf files with default params (see newParams).
func New(conf string) *Conf {
	return &Conf{
		generic.New(conf, newParams()),
	}
}

// Open opens and reads configuration from a file.
func Open(filename string) (*Conf, error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("could not read file %s: %s", filename, err)
	}
	conf := string(content)
	return New(conf), nil
}

// OpenReader reads configuration from a reader.
func OpenReader(r io.Reader) (*Conf, error) {
	content, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("could not read configuration from reader: %s", err)
	}
	conf := string(content)
	return New(conf), nil
}

// LookupFirst searches for a line that contains the given key, and if found,
// returns a Row structure for that line.
func (c *Conf) LookupFirst(keyCol int, key string) (*generic.Row, error) {
	row, _, err := c.LookupRow(keyCol, key, true, 0)
	if err != nil {
		return nil, err
	}
	return row, nil
}

// LookupAll searches for all rows that contains the given column value.
// Searching for values is case insensitive.
func (c *Conf) LookupAll(keyCol int, key string) ([]*generic.Row, error) {
	var rows []*generic.Row
	var offset int = 0
	for {
		// Find next row that has the key value
		r, nextOffset, err := c.LookupRow(keyCol, key, true, offset)
		if err != nil {
			break
		}
		rows = append(rows, r)
		offset = nextOffset
	}
	if len(rows) == 0 {
		return nil, generic.ErrKeyNotFound
	}
	return rows, nil
}

// Append adds a new row with the given values and returns a Row structure describing the
// line appended.
func (c *Conf) Append(connType, database, user, address, method string) (*Row, error) {
	c.EnsureEndsWithEOL()

	line := ""
	for i, val := range values {
		if i > 0 {
			line += c.params.DefaultDelim
		}
		line += val
	}

	writePos := len(c.conf)
	row, err := c.parseLine(line, writePos)
	if err != nil {
		return nil, errors.New("FAILED to parse the line that was about to be appended")
	}
	c.conf += line
	return row, nil
}
