package conf

import (
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	"github.com/quasoft/pgconf/generic"
)

const (
	keyCol   = 0
	valueCol = 1
)

// ErrKeyWithoutValue is returned if the key being looked up is found, but has no value.
// Keys with no values are not supported and the line containing it is ignored.
var ErrKeyWithoutValue = fmt.Errorf("key without value")

// NewParams creates param structure with defaults suitable for parsing of postgresql.conf files:
//  - Whitespace:             space, tab, carriage return and equal sign
//  - DefaultDelim: 		  =
//  - Quotes:                 " and '
//  - BackslashEscapedQuotes: true (allows use of \" and \' for escaping of quote characters in values)
//  - DefaultQuote:           '
//  - InlineComment:          #
//  - AlwaysQuoteStrings:	  true
func NewParams() generic.Params {
	return generic.Params{
		Whitespace:             " \t\r=",
		DefaultDelim:           " = ",
		Quotes:                 `"'`,
		BackslashEscapedQuotes: true,
		DefaultQuote:           '\'',
		InlineComment:          '#',
		AlwaysQuoteStrings:     true,
	}
}

// Conf represents a PostgreSQL configuration file (postgresql.conf).
type Conf struct {
	*generic.Conf
}

// New creates a new structure for reading/writing to postgresql.conf files with default params (see newParams).
func New(conf string) *Conf {
	return &Conf{
		generic.New(conf, NewParams()),
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

// LookupKey searches for a line that contains the given key, and if found,
// returns a Row structure for that line.
func (c *Conf) LookupKey(key string) (*generic.Row, error) {
	var row *generic.Row
	var offset int = 0
	for {
		// Find the last key that has any value
		r, nextOffset, err := c.LookupRow(keyCol, key, true, offset)
		if err != nil {
			break
		}
		if r.HasColumn(valueCol) {
			row = r
		}
		offset = nextOffset
	}
	if row == nil {
		return nil, generic.ErrKeyNotFound
	}
	return row, nil
}

// LookupOrAppendK searches for a line that contains the given key, and if found,
// returns a Row structure for that line.
// If not found, a new row is created and appended with an empty value.
// Searching for the key is case insensitive.
func (c *Conf) LookupOrAppendK(key string) (*generic.Row, error) {
	row, err := c.LookupKey(key)
	if err == generic.ErrKeyNotFound {
		return c.Append([]string{key, "''"}...)
	}
	if err != nil {
		return nil, err
	}
	return row, nil
}

// RawK retrieves the raw value of the key, including any quotes.
func (c *Conf) RawK(key string) (string, error) {
	row, err := c.LookupKey(key)
	if err != nil {
		return "", err
	}

	value, err := c.Raw(row, valueCol)
	if err != nil {
		return "", ErrKeyWithoutValue
	}

	return value, nil
}

// StringK retrieves the value of the key as a dequoted string.
// Removes the enclosing single quotes ('syslog' becomes just syslog),
// unescapes doubled quoted ('''users''') and backslash-quoted ('\'users\'')
// values.
func (c *Conf) StringK(key string) (string, error) {
	row, err := c.LookupKey(key)
	if err != nil {
		return "", err
	}

	return c.String(row, valueCol)
}

// IntK retrieves the value of the key as a dequoted integer.
func (c *Conf) IntK(key string) (int, error) {
	row, err := c.LookupKey(key)
	if err != nil {
		return 0, err
	}

	return c.Int(row, valueCol)
}

// Int64K retrieves the value of the key as a dequoted int64.
func (c *Conf) Int64K(key string) (int64, error) {
	row, err := c.LookupKey(key)
	if err != nil {
		return 0, err
	}

	return c.Int64(row, valueCol)
}

// Float64K retrieves the value of the key as a dequoted floating point number.
func (c *Conf) Float64K(key string) (float64, error) {
	row, err := c.LookupKey(key)
	if err != nil {
		return 0, err
	}

	return c.Float64(row, valueCol)
}

// BoolK retrieves the value of the key as a boolean.
// Values are expected to be one of: on, off, true, false, yes, no, 1, 0,
// or any unambiguous prefix of one of these.
// Case does not matter.
func (c *Conf) BoolK(key string) (bool, error) {
	value, err := c.StringK(key) // Read as string first to dequote the value
	if err != nil {
		return false, err
	}
	value = strings.TrimSpace(value)
	value = strings.ToLower(value)
	if value == "on" {
		return true, nil
	} else if strings.HasPrefix(value, "of") {
		return false, nil
	} else if strings.HasPrefix(value, "t") {
		return true, nil
	} else if strings.HasPrefix(value, "f") {
		return false, nil
	} else if strings.HasPrefix(value, "y") {
		return true, nil
	} else if strings.HasPrefix(value, "n") {
		return false, nil
	} else if value == "1" {
		return true, nil
	} else if value == "0" {
		return false, nil
	}
	return false, fmt.Errorf("unknown boolean value for key %s", key)
}

// SetRawK replaces the raw value of the specified key (including any quotes).
func (c *Conf) SetRawK(key string, value string) error {
	row, err := c.LookupOrAppendK(key)
	if err != nil {
		return err
	}
	return c.SetRaw(row, valueCol, value)
}

// SetStringK replaces the value of the specified key, enclosing it in single quotes.
func (c *Conf) SetStringK(key string, value string) error {
	row, err := c.LookupOrAppendK(key)
	if err != nil {
		return err
	}
	return c.SetString(row, valueCol, value)
}

// SetIntK replaces the value of the specified key with an unquoted integer value.
func (c *Conf) SetIntK(key string, value int) error {
	row, err := c.LookupOrAppendK(key)
	if err != nil {
		return err
	}
	return c.SetInt(row, valueCol, value)
}

// SetInt64K replaces the value of the specified key with an unquoted int64 value.
func (c *Conf) SetInt64K(key string, value int64) error {
	row, err := c.LookupOrAppendK(key)
	if err != nil {
		return err
	}
	return c.SetInt64(row, valueCol, value)
}

// SetFloat64K replaces the value of the specified key with a floating point number,
// while preserving whitespace on line.
// Outputs a string with the smallest number of digits needed to represent the value.
// If you want precision of your choice, or to enclose the value in quotes, use SetRawK instead.
func (c *Conf) SetFloat64K(key string, value float64) error {
	row, err := c.LookupOrAppendK(key)
	if err != nil {
		return err
	}
	return c.SetFloat64(row, valueCol, value)
}

// SetTrueFalseK replaces the value of the specified key with true or false.
func (c *Conf) SetTrueFalseK(key string, value bool) error {
	var raw string
	if value {
		raw = "true"
	} else {
		raw = "false"
	}
	return c.SetRawK(key, raw)
}

// SetOnOffK replaces the value of the specified key with on or off.
func (c *Conf) SetOnOffK(key string, value bool) error {
	var raw string
	if value {
		raw = "on"
	} else {
		raw = "off"
	}
	return c.SetRawK(key, raw)
}

// SetYesNoK replaces the value of the specified key with yes or no.
func (c *Conf) SetYesNoK(key string, value bool) error {
	var raw string
	if value {
		raw = "yes"
	} else {
		raw = "no"
	}
	return c.SetRawK(key, raw)
}
