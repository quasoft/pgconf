package conf_test

import (
	"bufio"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/quasoft/pgconf/conf"
)

func openTestFile(t *testing.T, testFile string) *conf.Conf {
	filename := filepath.Join("testdata", testFile)
	conf, err := conf.Open(filename)
	if err != nil {
		t.Fatalf(`Open("testdata/%s") failed: %s`, testFile, err)
	}
	if conf == nil {
		t.Fatalf(`Open("testdata/%s") = nil, want not nil`, testFile)
	}
	return conf
}

func openConfFile(t *testing.T) *conf.Conf {
	return openTestFile(t, "postgresql.conf")
}

func openEmptyFile(t *testing.T) *conf.Conf {
	return openTestFile(t, "empty.conf")
}

func readTestFile(t *testing.T, testFile string) string {
	filename := filepath.Join("testdata", testFile)
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		t.Fatalf(`Open("testdata/%s") failed: %s`, testFile, err)
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
	filename := filepath.Join("testdata", "thereisnosuchfile.conf")
	_, err := conf.Open(filename)
	if err == nil {
		t.Errorf(`Open("testdata/thereisnosuchfile.conf") should have failed with error`)
	}
}

func TestOpenReader(t *testing.T) {
	filename := filepath.Join("testdata", "postgresql.conf")
	f, err := os.Open(filename)
	if err != nil {
		t.Fatalf(`Open("testdata/postgresql.conf") failed: %s`, err)
	}
	conf, err := conf.OpenReader(f)
	if err != nil {
		t.Fatalf(`OpenReader() failed: %s`, err)
	}
	if conf == nil {
		t.Fatalf(`OpenReader("testdata/postgresql.conf") = nil, want not nil`)
	}
}

func TestRawK(t *testing.T) {
	conf := openConfFile(t)

	tests := []struct {
		name    string
		key     string
		want    string
		noerror bool
	}{
		{"Nonexisting key", "there_is_no_such_key", "", false},
		{"Key without value", "invalid_key_without_value", "", false},
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
			got, err := conf.RawK(tt.key)
			if err != nil && tt.noerror {
				t.Errorf("RawK(%q) errored with '%s', wanted no error", tt.key, err)
			} else if err == nil && !tt.noerror {
				t.Errorf("RawK(%q) did not error, wanted error", tt.key)
			} else if got != tt.want {
				t.Errorf("RawK(%q) = %q, want %q", tt.key, got, tt.want)
			}
		})
	}
}

func TestStringK(t *testing.T) {
	conf := openConfFile(t)

	tests := []struct {
		name    string
		key     string
		want    string
		noerror bool
	}{
		{"Nonexisting key", "there_is_no_such_key", "", false},
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
			got, err := conf.StringK(tt.key)
			if err != nil && tt.noerror {
				t.Errorf("StringK(%q) errored with '%s', wanted no error", tt.key, err)
			} else if err == nil && !tt.noerror {
				t.Errorf("StringK(%q) did not error, wanted error", tt.key)
			} else if got != tt.want {
				t.Errorf("StringK(%q) = %q, want %q", tt.key, got, tt.want)
			}
		})
	}
}

func TestIntK(t *testing.T) {
	conf := openConfFile(t)

	tests := []struct {
		name    string
		key     string
		want    int
		noerror bool
	}{
		{"Nonexisting key", "there_is_no_such_key", 0, false},
		{"Integer", "port", 5432, true},
		{"Quoted integer", "max_wal_senders", 10, true},
		{"Size", "shared_buffers", -1, false},
		{"Text value", "log_destination", -1, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := conf.IntK(tt.key)
			if err != nil && tt.noerror {
				t.Errorf("IntK(%q) errored with '%s', wanted no error", tt.key, err)
			} else if err == nil && !tt.noerror {
				t.Errorf("IntK(%q) did not error, wanted error", tt.key)
			} else if err == nil && got != tt.want {
				t.Errorf("IntK(%q) = %d, want %d", tt.key, got, tt.want)
			}
		})
	}
}

func TestInt64K(t *testing.T) {
	conf := openConfFile(t)

	tests := []struct {
		name    string
		key     string
		want    int64
		noerror bool
	}{
		{"Nonexisting key", "there_is_no_such_key", 0, false},
		{"Very large number", "autovacuum_multixact_freeze_max_age", 400000000000, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := conf.Int64K(tt.key)
			if err != nil && tt.noerror {
				t.Errorf("Int64K(%q) errored with '%s', wanted no error", tt.key, err)
			} else if err == nil && !tt.noerror {
				t.Errorf("Int64K(%q) did not error, wanted error", tt.key)
			} else if err == nil && got != tt.want {
				t.Errorf("Int64K(%q) = %d, want %d", tt.key, got, tt.want)
			}
		})
	}
}

func TestBoolK(t *testing.T) {
	conf := openConfFile(t)

	tests := []struct {
		name    string
		key     string
		want    bool
		noerror bool
	}{
		{"Nonexisting key", "there_is_no_such_key", false, false},
		{"Yes with no equal sign", "log_connections", true, true},
		{"On value", "ssl", true, true},
		{"Prefix of Off", "db_user_namespace", false, true},
		{"1 instead of On", "password_encryption", true, true},
		{"0 instead of Off", "wal_log_hints", false, true},
		{"Prefix of No", "bonjour", false, true},
		{"Prefix of True", "fsync", true, true},
		{"Prefix of Yes", "full_page_writes", true, true},
		{"Prefix of False", "wal_compression", false, true},
		{"Integer", "max_wal_senders", false, false},
		{"Text value", "log_destination", false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := conf.BoolK(tt.key)
			if err != nil && tt.noerror {
				t.Errorf("BoolK(%q) errored with '%s', wanted no error", tt.key, err)
			} else if err == nil && !tt.noerror {
				t.Errorf("BoolK(%q) did not error, wanted error", tt.key)
			} else if err == nil && got != tt.want {
				t.Errorf("BoolK(%q) = %v, want %v", tt.key, got, tt.want)
			}
		})
	}
}

func TestFloat64K(t *testing.T) {
	conf := openConfFile(t)

	tests := []struct {
		name    string
		key     string
		want    float64
		noerror bool
	}{
		{"Nonexisting key", "there_is_no_such_key", 0, false},
		{"Floating point with one digit after decimal", "checkpoint_completion_target", 0.5, true},
		{"Floating point with 4 digits after decimal", "cpu_operator_cost", 0.0025, true},
		{"Quoted integer", "max_wal_senders", 10, true},
		{"Integer", "max_connections", 100, true},
		{"Boolean", "log_connections", 0, false},
		{"Text value", "log_destination", 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := conf.Float64K(tt.key)
			if err != nil && tt.noerror {
				t.Errorf("Float64K(%q) errored with '%s', wanted no error", tt.key, err)
			} else if err == nil && !tt.noerror {
				t.Errorf("Float64K(%q) did not error, wanted error", tt.key)
			} else if err == nil && got != tt.want {
				t.Errorf("Float64K(%q) = %v, want %v", tt.key, got, tt.want)
			}
		})
	}
}

// TestSetRaw_EmptyFile tests if setting values in an empty config file appends new lines
// to the configuration file
func TestSetRawK_EmptyFile(t *testing.T) {
	wantContent := readTestFile(t, "postgresql-created.conf")
	conf := openEmptyFile(t)
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
			err := conf.SetRawK(tt.key, tt.value)
			if err != nil {
				t.Errorf("SetRawK(%q) errored with '%s', wanted no error", tt.key, err)
			} else {
				content := conf.All()
				got := readLine(t, content, tt.lineNumber)
				want := readLine(t, wantContent, tt.lineNumber)
				if got != want {
					t.Errorf("SetRawK(%q, %q) = got line #%d = %q, want %q", tt.key, tt.value, tt.lineNumber, got, want)
				}
			}
		})
	}
}

// TestSetRaw_ModifyExisting tests if updating an existing value preserves whitespace
// and comments on the same line
func TestSetRawK_ModifyExisting(t *testing.T) {
	wantContent := readTestFile(t, "postgresql-updated.conf")
	conf := openConfFile(t)
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
			err := conf.SetRawK(tt.key, tt.value)
			content := conf.All()
			got := readLine(t, content, tt.lineNumber)
			want := readLine(t, wantContent, tt.lineNumber)
			if err != nil {
				t.Errorf("SetRawK(%q) errored with '%s', wanted no error", tt.key, err)
			} else if got != want {
				t.Errorf("SetRawK(%q, %q) got line #%d = %q, want %q", tt.key, tt.value, tt.lineNumber, got, want)
			}
		})
	}
}

func TestSetStringK(t *testing.T) {
	wantContent := readTestFile(t, "postgresql-updated.conf")
	conf := openConfFile(t)
	tests := []struct {
		name       string
		key        string
		value      string
		lineNumber int // Line number in both postgresql.conf and postgresql-updated.conf test files
	}{
		{"String1", "listen_addresses", "127.0.0.1", 2},
		{"String2", "log_destination", "winevt", 8},
		{"Quoted string", "search_path", `"public"`, 10},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := conf.SetStringK(tt.key, tt.value)
			if err != nil {
				t.Errorf("SetStringK(%q) errored with '%s', wanted no error", tt.key, err)
			} else {
				content := conf.All()
				got := readLine(t, content, tt.lineNumber)
				want := readLine(t, wantContent, tt.lineNumber)
				if got != want {
					t.Errorf("SetStringK(%q, %q) = got line #%d = %q, want %q", tt.key, tt.value, tt.lineNumber, got, want)
				}
			}
		})
	}
}

func TestSetIntK(t *testing.T) {
	wantContent := readTestFile(t, "postgresql-updated.conf")
	conf := openConfFile(t)
	tests := []struct {
		name       string
		key        string
		value      int
		lineNumber int // Line number in both postgresql.conf and postgresql-updated.conf test files
	}{
		{"Int1", "port", 1234, 4},
		{"Int2", "max_connections", 9999, 5},
		{"Int3", "max_wal_senders", 20, 18},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := conf.SetIntK(tt.key, tt.value)
			if err != nil {
				t.Errorf("SetIntK(%q) errored with '%s', wanted no error", tt.key, err)
			} else {
				content := conf.All()
				got := readLine(t, content, tt.lineNumber)
				want := readLine(t, wantContent, tt.lineNumber)
				if got != want {
					t.Errorf("SetIntK(%q, %q) = got line #%d = %q, want %q", tt.key, tt.value, tt.lineNumber, got, want)
				}
			}
		})
	}
}

func TestSetInt64K(t *testing.T) {
	wantContent := readTestFile(t, "postgresql-updated.conf")
	conf := openConfFile(t)
	tests := []struct {
		name       string
		key        string
		value      int64
		lineNumber int // Line number in both postgresql.conf and postgresql-updated.conf test files
	}{
		{"Int64", "autovacuum_multixact_freeze_max_age", 512345678901, 37},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := conf.SetInt64K(tt.key, tt.value)
			if err != nil {
				t.Errorf("SetInt64K(%q) errored with '%s', wanted no error", tt.key, err)
			} else {
				content := conf.All()
				got := readLine(t, content, tt.lineNumber)
				want := readLine(t, wantContent, tt.lineNumber)
				if got != want {
					t.Errorf("SetInt64K(%q, %q) = got line #%d = %q, want %q", tt.key, tt.value, tt.lineNumber, got, want)
				}
			}
		})
	}
}

func TestSetFloat64K(t *testing.T) {
	wantContent := readTestFile(t, "postgresql-updated.conf")
	conf := openConfFile(t)
	tests := []struct {
		name       string
		key        string
		value      float64
		lineNumber int // Line number in both postgresql.conf and postgresql-updated.conf test files
	}{
		{"Float64-1", "checkpoint_completion_target", 0.1, 31},
		{"Float64-2", "cpu_operator_cost", 0.005, 32},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := conf.SetFloat64K(tt.key, tt.value)
			if err != nil {
				t.Errorf("SetFloat64K(%q) errored with '%s', wanted no error", tt.key, err)
			} else {
				content := conf.All()
				got := readLine(t, content, tt.lineNumber)
				want := readLine(t, wantContent, tt.lineNumber)
				if got != want {
					t.Errorf("SetFloat64K(%q, %f) = got line #%d = %q, want %q", tt.key, tt.value, tt.lineNumber, got, want)
				}
			}
		})
	}
}

func TestSetTrueFalseK(t *testing.T) {
	wantContent := readTestFile(t, "postgresql-updated.conf")
	conf := openConfFile(t)
	tests := []struct {
		name       string
		key        string
		value      bool
		lineNumber int // Line number in both postgresql.conf and postgresql-updated.conf test files
	}{
		{"Bool-1", "fsync", false, 25},
		{"Bool-2", "wal_compression", true, 27},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := conf.SetTrueFalseK(tt.key, tt.value)
			if err != nil {
				t.Errorf("SetTrueFalseK(%q) errored with '%s', wanted no error", tt.key, err)
			} else {
				content := conf.All()
				got := readLine(t, content, tt.lineNumber)
				want := readLine(t, wantContent, tt.lineNumber)
				if got != want {
					t.Errorf("SetTrueFalseK(%q, %v) = got line #%d = %q, want %q", tt.key, tt.value, tt.lineNumber, got, want)
				}
			}
		})
	}
}

func TestSetOnOffK(t *testing.T) {
	wantContent := readTestFile(t, "postgresql-updated.conf")
	conf := openConfFile(t)
	tests := []struct {
		name       string
		key        string
		value      bool
		lineNumber int // Line number in both postgresql.conf and postgresql-updated.conf test files
	}{
		{"Bool-1", "db_user_namespace", false, 22},
		{"Bool-2", "password_encryption", true, 23},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := conf.SetOnOffK(tt.key, tt.value)
			if err != nil {
				t.Errorf("SetOnOffK(%q) errored with '%s', wanted no error", tt.key, err)
			} else {
				content := conf.All()
				got := readLine(t, content, tt.lineNumber)
				want := readLine(t, wantContent, tt.lineNumber)
				if got != want {
					t.Errorf("SetOnOffK(%q, %v) = got line #%d = %q, want %q", tt.key, tt.value, tt.lineNumber, got, want)
				}
			}
		})
	}
}

func TestSetYesNoK(t *testing.T) {
	wantContent := readTestFile(t, "postgresql-updated.conf")
	conf := openConfFile(t)
	tests := []struct {
		name       string
		key        string
		value      bool
		lineNumber int // Line number in both postgresql.conf and postgresql-updated.conf test files
	}{
		{"Bool-1", "full_page_writes", false, 26},
		{"Bool-2", "wal_log_hints", true, 28},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := conf.SetYesNoK(tt.key, tt.value)
			if err != nil {
				t.Errorf("SetYesNoK(%q) errored with '%s', wanted no error", tt.key, err)
			} else {
				content := conf.All()
				got := readLine(t, content, tt.lineNumber)
				want := readLine(t, wantContent, tt.lineNumber)
				if got != want {
					t.Errorf("SetYesNoK(%q, %v) = got line #%d = %q, want %q", tt.key, tt.value, tt.lineNumber, got, want)
				}
			}
		})
	}
}
