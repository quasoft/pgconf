package pgconf

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
)

// File represents a PostgreSQL configuration file (postgresql.conf).
type File struct {
	ConfFile   string
	Whitespace string
	data       string
}

// newFile creates a new File structure with default whitespace settings (space, tab and return).
func newFile(confFile string) *File {
	return &File{
		ConfFile:   confFile,
		Whitespace: " \t\r",
	}
}

// Open opens and reads into memory an existing configuration file.
func Open(confFile string) (*File, error) {
	bytes, err := ioutil.ReadFile(confFile)
	if err != nil {
		return nil, err
	}
	conf := newFile(confFile)
	conf.data = string(bytes)
	return conf, nil
}

// Save stores to disk the changes made to the in-memory configuration.
func (f *File) Save() error {
	return ioutil.WriteFile(f.ConfFile, []byte(f.data), 0)
}

// findParam performs a case-insensitive search for the given key name,
// and returns a param structure with the positions to the start and end of
// the key name and the value.
// Positions are used internally to preserve whitespace while updating values.
func (f *File) findParam(key string) (*param, error) {
	var offset int64
	r := bufio.NewReader(strings.NewReader(f.data))
	for {
		line, errRead := r.ReadString('\n')

		searchLine := strings.TrimLeft(line, f.Whitespace)
		searchLine = strings.ToLower(searchLine)
		if strings.HasPrefix(searchLine, strings.ToLower(key)) {
			p, err := f.parseLine(line, offset)
			if err == nil {
				return p, err
			}
		}

		if errRead != nil {
			break
		}

		offset = offset + int64(len(line))
	}
	return nil, ErrKeyNotFound
}

// parseLine scans the given line and returns a param structure with the start and end positions
// of the key name and the value. Positions are relative to the start of the file and do not include
// whitespace.
func (f *File) parseLine(line string, offset int64) (p *param, err error) {
	p = newParam()
	err = nil

	var pos int64 = -1
	var insideQuote bool
	var lastRune rune
	for i, r := range line {
		// Stop on comment or line ending
		if r == '#' || r == '\n' {
			break
		}

		pos = offset + int64(i)

		isWhitespace := strings.Index(f.Whitespace, string(r)) > -1
		isQuote := strings.Index(`'`, string(r)) > -1
		isEqual := (r == '=')

		if p.value.start > -1 {
			if isQuote && lastRune != '\\' {
				insideQuote = !insideQuote
			}
			if isWhitespace && !insideQuote {
				p.value.end = pos
				break
			}
		} else if p.key.end > -1 {
			if isWhitespace || isEqual {
				continue
			}
			p.value.start = pos
			if isQuote {
				insideQuote = true
			}
		} else if p.key.start > -1 {
			if isWhitespace || isEqual {
				p.key.end = pos
			}
		} else if p.key.start == -1 {
			if isWhitespace {
				continue
			}
			p.key.start = pos
		}

		lastRune = r
	}

	if p.value.start > -1 && p.value.end == -1 {
		p.value.end = pos + 1
	} else if p.key.start > -1 && p.key.end == -1 {
		p.key.end = pos + 1
	}

	if p.key.end > -1 && p.value.start == -1 {
		err = ErrKeyWithoutValue
	} else if p.key.start == -1 {
		err = ErrEmptyLine
	}

	return
}

// Raw retrieves the raw value of the key, including any quotes.
func (f *File) Raw(key string) (string, error) {
	p, err := f.findParam(key)
	if err != nil {
		return "", err
	}

	offset := p.value.start
	size := p.valueSize()
	if size <= 0 {
		return "", fmt.Errorf("value size is unknown")
	}

	value := f.data[offset : offset+size]

	return value, nil
}

// AsString retrieves the value of the key as a dequoted string.
// Removes the enclosing single quotes ('syslog' becomes just syslog),
// unescapes doubled quoted ('''users''') and backslash-quoted ('\'users\'')
// values.
func (f *File) AsString(key string) (string, error) {
	value, err := f.Raw(key)
	if err != nil {
		return "", err
	}

	start := 0
	end := len(value)
	if strings.HasPrefix(value, "'") {
		start++
	}
	if strings.HasSuffix(value, "'") {
		end--
	}
	value = value[start:end]
	value = strings.Replace(value, "''", "'", -1)
	value = strings.Replace(value, `\'`, "'", -1)
	return value, nil
}

// AsInt retrieves the value of the key as a dequoted integer.
func (f *File) AsInt(key string) (int, error) {
	value, err := f.AsString(key) // Read as string first to dequote the value
	if err != nil {
		return -1, err
	}
	value = strings.TrimSpace(value)
	return strconv.Atoi(value)
}

// AsInt64 retrieves the value of the key as a dequoted int64.
func (f *File) AsInt64(key string) (int64, error) {
	value, err := f.AsString(key) // Read as string first to dequote the value
	if err != nil {
		return -1, err
	}
	value = strings.TrimSpace(value)
	return strconv.ParseInt(value, 10, 64)
}

// AsBool retrieves the value of the key as a boolean.
// Values are expected to be one of: on, off, true, false, yes, no, 1, 0,
// or any unambiguous prefix of one of these.
// Case does not matter.
func (f *File) AsBool(key string) (bool, error) {
	value, err := f.AsString(key) // Read as string first to dequote the value
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

// AsFloat64 retrieves the value of the key as a dequoted floating point number.
func (f *File) AsFloat64(key string) (float64, error) {
	value, err := f.AsString(key) // Read as string first to dequote the value
	if err != nil {
		return -1, err
	}
	value = strings.TrimSpace(value)
	return strconv.ParseFloat(value, 64)
}

// SetRaw replaces the raw value of the specified key (including any quotes).
func (f *File) SetRaw(key string, value string) error {
	p, err := f.findParam(key)
	if err == ErrKeyNotFound {
		line := fmt.Sprintf("%s = %s", key, value)
		if len(f.data) > 0 && f.data[len(f.data)-1] != '\n' {
			line = "\n" + line
		}
		f.data += line
		return nil
	}
	if err != nil {
		return err
	}

	offset := p.value.start
	oldSize := p.valueSize()
	if oldSize <= 0 {
		return fmt.Errorf("value size is unknown")
	}

	f.data = f.data[:offset] + value + f.data[offset+oldSize:]
	return nil
}

// SetString replaces the value of the specified key, enclosing it in single quotes.
func (f *File) SetString(key string, value string) error {
	value = strings.Replace(value, "'", "''", -1)
	raw := "'" + value + "'"
	return f.SetRaw(key, raw)
}

// SetInt replaces the value of the specified key with an unquoted integer value.
func (f *File) SetInt(key string, value int) error {
	raw := strconv.Itoa(value)
	return f.SetRaw(key, raw)
}

// SetBool replaces the value of the specified key with true or false
func (f *File) SetBool(key string, value bool) error {
	raw := strconv.FormatBool(value)
	return f.SetRaw(key, raw)
}

// SetOnOff replaces the value of the specified key with on or off.
func (f *File) SetOnOff(key string, value bool) error {
	var raw string
	if value {
		raw = "on"
	} else {
		raw = "off"
	}
	return f.SetRaw(key, raw)
}

// SetYesNo replaces the value of the specified key with yes or no.
func (f *File) SetYesNo(key string, value bool) error {
	var raw string
	if value {
		raw = "yes"
	} else {
		raw = "no"
	}
	return f.SetRaw(key, raw)
}

// SetFloat64 replaces the value of the specified key with a floating point number.
// Internally uses fmt.Sprintf("%f") for formatiing. If you want to write a float with
// precision of your choice, format it yourself and write it with the SetRaw function.
func (f *File) SetFloat64(key string, value float64) error {
	raw := fmt.Sprintf("%f", value)
	return f.SetRaw(key, raw)
}

// ReadAll returns the whole configuration file as a string
func (f *File) ReadAll() string {
	return f.data
}
