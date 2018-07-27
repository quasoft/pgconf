package generic

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

// Params allows the caller to customize the behaviour of generic.Conf.
type Params struct {
	Whitespace             string // One or more characters that should be recognized as whitespace
	DefaultDelim           string // A sequence of whitespace characters to use as delimiter when appending new rows
	Quotes                 string // One or more characters that should be recognized as value quoating character
	BackslashEscapedQuotes bool   // Whether backslash escaped quotes like \' or \" should be recognized
	DefaultQuote           rune   // Default quote to use when updating or adding new string values
	InlineComment          rune   // Character that denotes inline comments (usually # or ;)
	AlwaysQuoteStrings     bool   // If true string values are enclosed in quotes even if the values contain no quotes
}

// NewParams creates a new configuration with the following defaults:
//  - Whitespace:             space, tab and carriage return
//  - DefaultDelim: 		  tab
//  - Quotes:                 " and '
//  - BackslashEscapedQuotes: true (allows use of \" and \' for escaping of quote characters in values)
//  - DefaultQuote:           "
//  - InlineComment:          #
//  - AlwaysQuoteStrings:	  false
func NewParams() Params {
	return Params{
		Whitespace:             " \t\r",
		DefaultDelim:           "\t",
		Quotes:                 `"'`,
		BackslashEscapedQuotes: true,
		DefaultQuote:           '"',
		InlineComment:          '#',
		AlwaysQuoteStrings:     false,
	}
}

// Conf can read and write to multi-column whitespace delimited configurations, while preserving existing
// whitespace, when updating values.
type Conf struct {
	conf   string
	params Params
}

// New creates a new conf structure for reading/writing to the specified configuration.
func New(conf string, params Params) *Conf {
	return &Conf{
		conf:   conf,
		params: params,
	}
}

// Open opens and reads configuration from a file.
func Open(filename string, params Params) (*Conf, error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("could not read file %s: %s", filename, err)
	}
	conf := string(content)
	return New(conf, params), nil
}

// OpenReader reads configuration from a reader.
func OpenReader(r io.Reader, params Params) (*Conf, error) {
	content, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("could not read configuration from reader: %s", err)
	}
	conf := string(content)
	return New(conf, params), nil
}

// SetParams updates the parameters that determine the behaviour of generic.Conf.
func (c *Conf) SetParams(params Params) {
	c.params = params
}

// All returns the whole configuration as a string.
func (c *Conf) All() string {
	return c.conf
}

// WriteTo writes the whole configuration to a writer.
func (c *Conf) WriteTo(w io.Writer) (int64, error) {
	n, err := io.WriteString(w, c.conf)
	return int64(n), err
}

// WriteFile writes the whole configuration to a file.
func (c *Conf) WriteFile(filename string, perm os.FileMode) error {
	return ioutil.WriteFile(filename, []byte(c.conf), perm)
}

// LookupRow searches for a row that contains the given column value starting at offset,
// and if found, returns a Row structure with start and end positions of all column values and
// the offset position of the next line.
// on the row. Positions are used internally to preserve whitespace when updating values.
// If ignoreCase is true a case insensitive search is performed.
func (c *Conf) LookupRow(keyCol int, key string, ignoreCase bool, offset int) (*Row, int, error) {
	str := strings.NewReader(c.conf[offset:])
	reader := bufio.NewReader(str)
	for {
		line, errRead := reader.ReadString('\n')
		endOfLine := offset + len(line)

		row, err := c.parseLine(line, offset)
		if err == nil && row.HasColumn(keyCol) {
			rowKey, err := c.Raw(row, keyCol)
			if err == nil &&
				rowKey == key ||
				(ignoreCase && strings.ToLower(rowKey) == strings.ToLower(key)) {
				return row, endOfLine, nil
			}
		}

		if errRead != nil {
			break
		}

		offset = endOfLine
	}
	return nil, 0, ErrKeyNotFound
}

// parseLine scans the given line and returns a param structure with the start and end positions
// of the key name and the value. Positions are relative to the start of the buffer/file and do not include
// whitespace.
func (c *Conf) parseLine(line string, offset int) (row *Row, err error) {
	row = newRow()
	err = nil

	var pos int = -1
	var insideQuote bool
	var expectedQuote = c.params.Quotes // Match any of the quote characters specified in params
	var lastRune rune
	var start, end int = -1, -1
	for i, r := range line {
		// Stop on inline comment or line ending
		if r == c.params.InlineComment || r == '\n' {
			break
		}

		pos = offset + i

		isWhitespace := strings.Index(c.params.Whitespace, string(r)) > -1
		isQuote := strings.Index(expectedQuote, string(r)) > -1

		if start > -1 {
			if isQuote && (!c.params.BackslashEscapedQuotes || lastRune != '\\') {
				insideQuote = !insideQuote
				if insideQuote {
					expectedQuote = string(r) // Quoted value can be closed only with exactly the same quote character
				} else {
					expectedQuote = c.params.Quotes // Quote can start with any quote
				}
			}
			if isWhitespace && !insideQuote {
				// This is the end if a column value, so add this token to the row,
				end = pos
				row.addToken(start, end)

				// and start parsing the next column value
				start = -1
				end = -1
			}
		} else {
			if isWhitespace {
				continue
			}
			// This is beginning of a column value
			start = pos
			if isQuote {
				insideQuote = true
				expectedQuote = string(r) // Quoted value can be closed only with exactly the same quote character
			}
		}

		lastRune = r
	}

	// Finialize the last column value
	if start > -1 && end == -1 {
		end = pos + 1
		row.addToken(start, end)
	}

	// Error on lines with whitespace and comments only
	if start == -1 && row.ColCount() == 0 {
		err = ErrEmptyLine
	}

	return
}

// HasQuotesOrWhitespace tests if the value contains any of the quote characters specified in Params.DefaultQuote
// or a whitespace character specified in Params.Whitespace.
func (c *Conf) HasQuotesOrWhitespace(value string) bool {
	for _, r := range value {
		ch := string(r)
		isWhitespace := strings.Index(c.params.Whitespace, ch) > -1
		isQuote := strings.Index(c.params.Quotes, ch) > -1

		if isWhitespace || isQuote {
			return true
		}
	}
	return false
}

// EscapeQuotes escapes quote characters by double-quoting them.
// Recognizes and double-quotes characters specified in Params.Quotes.
func (c *Conf) EscapeQuotes(value string, quote rune) string {
	single := string(quote)
	doubled := strings.Repeat(single, 2)
	value = strings.Replace(value, single, doubled, -1)
	return value
}

// UnescapeQuotes unescapes double quoted or backslash quoted values,
// depending on settings in Params.
func (c *Conf) UnescapeQuotes(value string, quote rune) string {
	single := string(quote)
	doubled := strings.Repeat(single, 2)
	value = strings.Replace(value, doubled, single, -1)

	if c.params.BackslashEscapedQuotes {
		escaped := `\` + single
		value = strings.Replace(value, escaped, single, -1)
	}
	return value
}

// Quote escapes any quotes in value by double-quoting them and then encloses the escaped value
// with the quote character specified in Params.DefaultQuote.
func (c *Conf) Quote(value string) string {
	quote := c.params.DefaultQuote
	return string(quote) + c.EscapeQuotes(value, quote) + string(quote)
}

// Dequote removes enclosing quotes and unescapes double quotes and backslash escaped quotes in values.
func (c *Conf) Dequote(value string) string {
	if len(value) < 2 {
		return value
	}

	first := value[:1]
	if strings.Index(c.params.Quotes, first) == -1 {
		// First char not a quote, just return value as it is
		return value
	}

	last := value[len(value)-1:]
	if last != first {
		// Last char is not the same quote as the first char, so just return value as it is
		return value
	}

	value = value[1 : len(value)-1]

	return c.UnescapeQuotes(value, rune(first[0]))
}

// Raw retrieves the raw value of the column at the specified row, including quotes, but excluding surrounding
// whitespace around quotes (if any).
func (c *Conf) Raw(row *Row, col int) (string, error) {
	if row == nil {
		return "", errors.New("could not retrieve raw value for a nil row")
	}
	token, err := row.Token(col)
	if err != nil {
		return "", fmt.Errorf("could not retrieve token for column %d: %s", col, err)
	}

	offset := token.Start
	size, err := token.ValueSize()
	if err != nil {
		return "", fmt.Errorf("could not get value size for column %d: %s", col, err)
	}
	if size <= 0 {
		return "", fmt.Errorf("got value size of %d for column %d, want size > 0", size, col)
	}

	raw := c.conf[offset : offset+size]

	return raw, nil
}

// String retrieves the value of the column at an existing row as a dequoted string.
// Removes the enclosing quotes and unescapes double quotes and backslashed quotes in value.
func (c *Conf) String(row *Row, col int) (string, error) {
	value, err := c.Raw(row, col)
	if err != nil {
		return "", err
	}

	return c.Dequote(value), nil
}

// Int retrieves the value of the column at an existing row as a dequoted integer.
func (c *Conf) Int(row *Row, col int) (int, error) {
	value, err := c.String(row, col) // Read as string first to dequote the value
	if err != nil {
		return -1, err
	}
	value = strings.TrimSpace(value)
	return strconv.Atoi(value)
}

// Int64 retrieves the value of the column at an existing row as a dequoted int64.
func (c *Conf) Int64(row *Row, col int) (int64, error) {
	value, err := c.String(row, col) // Read as string first to dequote the value
	if err != nil {
		return -1, err
	}
	value = strings.TrimSpace(value)
	return strconv.ParseInt(value, 10, 64)
}

// Float64 retrieves the value of the column at an existing row as a dequoted floating point number.
func (c *Conf) Float64(row *Row, col int) (float64, error) {
	value, err := c.String(row, col) // Read as string first to dequote the value
	if err != nil {
		return -1, err
	}
	value = strings.TrimSpace(value)
	return strconv.ParseFloat(value, 64)
}

// EnsureEndsWithEOL makes sure that the configuration ends with an EOL character,
// unless the configuration is empty.
func (c *Conf) EnsureEndsWithEOL() {
	if c.conf == "" {
		// Configuration is empty. There is no need to add an EOL character.
		return
	}

	if c.conf[len(c.conf)-1] == '\n' {
		// Already ends with EOL, do nothing
		return
	}

	c.conf += "\n"

	return
}

// Append adds a new row with the given column values and returns a Row structure describing the
// line appended.
func (c *Conf) Append(values ...string) (*Row, error) {
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

// SetRaw replaces the raw value of the column at an existing row, including any quotes,
// while preserving whitespace on the line.
func (c *Conf) SetRaw(row *Row, col int, value string) error {
	if row == nil {
		return errors.New("could not retrieve raw value for a nil row")
	}
	token, err := row.Token(col)
	if err != nil {
		return fmt.Errorf("could not retrieve token for column: %s", err)
	}

	offset := token.Start
	oldSize, err := token.ValueSize()
	if err != nil {
		return fmt.Errorf("could not get value size: %s", err)
	}
	if oldSize <= 0 {
		return fmt.Errorf("got value size of %d, want size > 0", oldSize)
	}

	c.conf = c.conf[:offset] + value + c.conf[offset+oldSize:]

	return nil
}

// SetString encloses the given value with single quotes and updates the existing value
// at the specified row and column, while preserving whitespace on line.
func (c *Conf) SetString(row *Row, col int, value string) error {
	var raw string
	if c.params.AlwaysQuoteStrings || c.HasQuotesOrWhitespace(value) {
		raw = c.Quote(value)
	} else {
		raw = value
	}

	return c.SetRaw(row, col, raw)
}

// SetInt updates the existing value at the specified row and column with an unquoted integer,
// while preserving whitespace on line.
func (c *Conf) SetInt(row *Row, col int, value int) error {
	raw := strconv.Itoa(value)
	return c.SetRaw(row, col, raw)
}

// SetInt64 updates the existing value at the specified row and column with an unquoted int64,
// while preserving whitespace on line.
func (c *Conf) SetInt64(row *Row, col int, value int64) error {
	raw := strconv.FormatInt(value, 10)
	return c.SetRaw(row, col, raw)
}

// SetFloat64 updates the existing value at the specified row and column with an unquoted
// floating point number, while preserving whitespace on line.
// Outputs a string with the smallest number of digits needed to represent the value.
// If you want precision of your choice, or to enclose the value in quotes, use SetRaw instead.
func (c *Conf) SetFloat64(row *Row, col int, value float64) error {
	raw := strconv.FormatFloat(value, 'f', -1, 64)
	return c.SetRaw(row, col, raw)
}
