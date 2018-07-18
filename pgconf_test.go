package pgconf_test

import (
	"bufio"
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"

	"github.com/quasoft/pgconf"
)

func openTestFile(t *testing.T, testFile string) *pgconf.File {
	path := filepath.Join("testdata", testFile)
	f, err := pgconf.Open(path)
	if err != nil {
		t.Fatalf(`Open("testdata/postgresql.conf") failed: %s`, err)
	}

	if f == nil {
		t.Fatalf(`Open("testdata/postgresql.conf") = nil, want not nil`)
	}

	return f
}

func openConfFile(t *testing.T) *pgconf.File {
	return openTestFile(t, "postgresql.conf")
}

func openEmptyFile(t *testing.T) *pgconf.File {
	return openTestFile(t, "empty.conf")
}

func readTestFile(t *testing.T, testFile string) string {
	path := filepath.Join("testdata", testFile)
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatalf(`Open("testdata/postgresql.conf") failed: %s`, err)
	}
	return string(bytes)
}

// readLine reads a line of text by its line number
func readLine(t *testing.T, fileContent string, line int) string {
	r := strings.NewReader(fileContent)
	s := bufio.NewScanner(r)
	l := 0
	for s.Scan() {
		l++
		if l == line {
			return s.Text()
		}
	}
	return ""
}

func TestOpen_NotExisting(t *testing.T) {
	path := filepath.Join("testdata", "thereisnosuchfile.conf")
	_, err := pgconf.Open(path)
	if err == nil {
		t.Fatalf(`Open("testdata/postgresql.conf") should have failed with error`)
	}
}

func TestRaw(t *testing.T) {
	f := openConfFile(t)

	tests := []struct {
		name    string
		key     string
		want    string
		noerror bool
	}{
		{"Quoted string value", "listen_addresses", "'*'", true},
		{"White space everywhere", "port", "5432", true},
		{"Only trailing whitespace", "max_connections", "100", true},
		{"No equal sign", "log_connections", "yes", true},
		{"No whitespace", "log_destination", "'syslog'", true},
		{"Quoted strings", "search_path", `'"$user", ''public'', \'other\''`, true},
		{"Key without value", "nosuchkey", "", false},
		{"No comment and no trailing whitespace", "shared_buffers", "128MB", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := f.Raw(tt.key)
			if err != nil && tt.noerror {
				t.Errorf("Raw(%q) errored with '%s', wanted no error", tt.key, err)
			} else if err == nil && !tt.noerror {
				t.Errorf("Raw(%q) did not error, wanted error", tt.key)
			} else if got != tt.want {
				t.Errorf("Raw(%q) = %q, want %q", tt.key, got, tt.want)
			}
		})
	}
}

func TestAsString(t *testing.T) {
	f := openConfFile(t)

	tests := []struct {
		name    string
		key     string
		want    string
		noerror bool
	}{
		{"Quoted string value", "listen_addresses", "*", true},
		{"White space everywhere", "port", "5432", true},
		{"Only trailing whitespace", "max_connections", "100", true},
		{"No equal sign", "log_connections", "yes", true},
		{"No whitespace", "log_destination", "syslog", true},
		{"Quoted strings", "search_path", `"$user", 'public', 'other'`, true},
		{"No comment and no trailing whitespace", "shared_buffers", "128MB", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := f.AsString(tt.key)
			if err != nil && tt.noerror {
				t.Errorf("Raw(%q) errored with '%s', wanted no error", tt.key, err)
			} else if err == nil && !tt.noerror {
				t.Errorf("Raw(%q) did not error, wanted error", tt.key)
			} else if got != tt.want {
				t.Errorf("Raw(%q) = %q, want %q", tt.key, got, tt.want)
			}
		})
	}
}

func TestAsInt(t *testing.T) {
	f := openConfFile(t)

	tests := []struct {
		name    string
		key     string
		want    int
		noerror bool
	}{
		{"Integer", "port", 5432, true},
		{"Quoted integer", "max_wal_senders", 10, true},
		{"Size", "shared_buffers", -1, false},
		{"Text value", "log_destination", -1, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := f.AsInt(tt.key)
			if err != nil && tt.noerror {
				t.Errorf("Raw(%q) errored with '%s', wanted no error", tt.key, err)
			} else if err == nil && !tt.noerror {
				t.Errorf("Raw(%q) did not error, wanted error", tt.key)
			} else if err == nil && got != tt.want {
				t.Errorf("Raw(%q) = %d, want %d", tt.key, got, tt.want)
			}
		})
	}
}

func TestAsInt64(t *testing.T) {
	f := openConfFile(t)

	tests := []struct {
		name    string
		key     string
		want    int
		noerror bool
	}{
		{"Very large number", "autovacuum_multixact_freeze_max_age", 400000000000, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := f.AsInt(tt.key)
			if err != nil && tt.noerror {
				t.Errorf("Raw(%q) errored with '%s', wanted no error", tt.key, err)
			} else if err == nil && !tt.noerror {
				t.Errorf("Raw(%q) did not error, wanted error", tt.key)
			} else if err == nil && got != tt.want {
				t.Errorf("Raw(%q) = %d, want %d", tt.key, got, tt.want)
			}
		})
	}
}

func TestAsBool(t *testing.T) {
	f := openConfFile(t)

	tests := []struct {
		name    string
		key     string
		want    bool
		noerror bool
	}{
		{"Yes with no equal sign", "log_connections", true, true},
		{"On value", "ssl", true, true},
		{"Prefix of Off", "db_user_namespace", false, true},
		{"1 instead of On", "password_encryption", true, true},
		{"Prefix of No", "bonjour", false, true},

		{"Integer", "max_wal_senders", false, false},
		{"Text value", "log_destination", false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := f.AsBool(tt.key)
			if err != nil && tt.noerror {
				t.Errorf("Raw(%q) errored with '%s', wanted no error", tt.key, err)
			} else if err == nil && !tt.noerror {
				t.Errorf("Raw(%q) did not error, wanted error", tt.key)
			} else if err == nil && got != tt.want {
				t.Errorf("Raw(%q) = %v, want %v", tt.key, got, tt.want)
			}
		})
	}
}

func TestAsFloat64(t *testing.T) {
	f := openConfFile(t)

	tests := []struct {
		name    string
		key     string
		want    float64
		noerror bool
	}{
		{"Floating point with one digit after decimal", "checkpoint_completion_target", 0.5, true},
		{"Floating point with 4 digits after decimal", "cpu_operator_cost", 0.0025, true},
		{"Quoted integer", "max_wal_senders", 10, true},
		{"Integer", "max_connections", 100, true},
		{"Boolean", "log_connections", 0, false},
		{"Text value", "log_destination", 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := f.AsFloat64(tt.key)
			if err != nil && tt.noerror {
				t.Errorf("Raw(%q) errored with '%s', wanted no error", tt.key, err)
			} else if err == nil && !tt.noerror {
				t.Errorf("Raw(%q) did not error, wanted error", tt.key)
			} else if err == nil && got != tt.want {
				t.Errorf("Raw(%q) = %v, want %v", tt.key, got, tt.want)
			}
		})
	}
}

// TestSetRaw_EmptyFile tests if setting values in an empty config file appends new lines
// to the configuration file
func TestSetRaw_EmptyFile(t *testing.T) {
	wantContent := readTestFile(t, "postgresql-created.conf")
	f := openEmptyFile(t)
	tests := []struct {
		name       string
		key        string
		value      string
		lineNumber int // Line number in the postgresql-created.conf test file (wanted lines)
	}{
		{"Number", "port", "5432", 1},
		{"Boolean", "log_connections", "yes", 2},
		{"Quoted string", "log_destination", "'syslog'", 3},
		{"Double quoted strings", "search_path", `'"$user", ''public'', \'other\''`, 4},
		{"Size", "shared_buffers", "128MB", 5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := f.SetRaw(tt.key, tt.value)
			if err != nil {
				t.Errorf("SetRaw(%q) errored with '%s', wanted no error", tt.key, err)
			} else {
				content := f.ReadAll()
				got := readLine(t, content, tt.lineNumber)
				want := readLine(t, wantContent, tt.lineNumber)
				if got != want {
					t.Errorf("SetRaw(%q, %q) = got line #%d = %q, want %q", tt.key, tt.value, tt.lineNumber, got, want)
				}
			}
		})
	}
}

// TestSetRaw_ModifyExisting tests if updating an existing value preserves whitespace
// and comments on the same line
func TestSetRaw_ModifyExisting(t *testing.T) {
	wantContent := readTestFile(t, "postgresql-updated.conf")
	f := openConfFile(t)
	tests := []struct {
		name       string // Name of test case
		key        string // Key to change
		value      string // Value to use
		lineNumber int    // Line number in both postgresql.conf and postgresql-updated.conf test files
	}{
		{"Quoted string", "listen_addresses", "'127.0.0.1'", 2},
		{"Number", "port", "1234", 4},
		{"Another number", "max_connections", "9999", 5},
		{"Boolean", "log_connections", "no", 7},
		{"Another quoted string", "log_destination", "'winevt'", 8},
		{"Double quoted strings", "search_path", `'"public"'`, 10},
		{"Size", "shared_buffers", "256MB", 15},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := f.SetRaw(tt.key, tt.value)
			content := f.ReadAll()
			got := readLine(t, content, tt.lineNumber)
			want := readLine(t, wantContent, tt.lineNumber)
			if err != nil {
				t.Errorf("SetRaw(%q) errored with '%s', wanted no error", tt.key, err)
			} else if got != want {
				t.Errorf("SetRaw(%q, %q) got line #%d = %q, want %q", tt.key, tt.value, tt.lineNumber, got, want)
			}
		})
	}
}
