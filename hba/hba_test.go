package hba_test

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/quasoft/pgconf/hba"
)

func openTestFile(t *testing.T, testFile string) *hba.Conf {
	filename := filepath.Join("testdata", testFile)
	conf, err := hba.Open(filename)
	if err != nil {
		t.Fatalf(`Open("testdata/%s") failed: %s`, testFile, err)
	}
	if conf == nil {
		t.Fatalf(`Open("testdata/%s") = nil, want not nil`, testFile)
	}
	return conf
}

func TestOpen_NotExisting(t *testing.T) {
	filename := filepath.Join("testdata", "thereisnosuchfile.conf")
	_, err := hba.Open(filename)
	if err == nil {
		t.Errorf(`Open("testdata/thereisnosuchfile.conf") should have failed with error`)
	}
}

func TestOpenReader(t *testing.T) {
	filename := filepath.Join("testdata", "sample.conf")
	f, err := os.Open(filename)
	if err != nil {
		t.Fatalf(`Open("testdata/sample.conf") failed: %s`, err)
	}
	conf, err := hba.OpenReader(f)
	if err != nil {
		t.Fatalf(`OpenReader() failed: %s`, err)
	}
	if conf == nil {
		t.Fatalf(`OpenReader("testdata/sample.conf") = nil, want not nil`)
	}
}

func TestLookupFirst(t *testing.T) {
	conf := openTestFile(t, "sample.conf")

	tests := []struct {
		name    string
		keyCol  int
		key     string
		noerror bool
	}{
		{"By ConnType", hba.ConnType, "host", true},
		{"By database", hba.Database, "replication", true},
		{"By user", hba.User, "postgres", true},
		{"By address", hba.Address, "::1/128", true},
		{"By method", hba.Method, "md5", true},
		{"Not existing", hba.Database, "nosuchdatabase", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			row, err := conf.LookupFirst(tt.keyCol, tt.key)
			if err != nil && tt.noerror {
				t.Errorf("LookupFirst(%d, %q) errored with '%s', wanted no error", tt.keyCol, tt.key, err)
			} else if err == nil && !tt.noerror {
				t.Errorf("LookupFirst(%d, %q) did not error, wanted error", tt.keyCol, tt.key)
			} else if tt.noerror && row == nil {
				t.Errorf("LookupFirst(%d, %q) got nil, wanted a row structure", tt.keyCol, tt.key)
			}
		})
	}
}

func TestLookupAll(t *testing.T) {
	conf := openTestFile(t, "sample.conf")

	tests := []struct {
		name        string
		keyCol      int
		key         string
		wantNumRows int
		noerror     bool
	}{
		{"1 row", hba.Address, "10.0.0.3/32", 1, true},
		{"2 rows", hba.Address, "127.0.0.1/32", 2, true},
		{"3 rows", hba.Database, "replication", 3, true},
		{"5 rows", hba.Method, "md5", 5, true},
		{"Not existing", hba.Database, "thereisnodatabase", 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rows, err := conf.LookupAll(tt.keyCol, tt.key)
			gotNumRows := len(rows)
			if err != nil && tt.noerror {
				t.Errorf("LookupAll(%d, %q) errored with '%s', wanted no error", tt.keyCol, tt.key, err)
			} else if err == nil && !tt.noerror {
				t.Errorf("LookupAll(%d, %q) did not error, wanted error", tt.keyCol, tt.key)
			} else if gotNumRows != tt.wantNumRows {
				t.Errorf("LookupAll(%d, %q) got %d rows, want %d", tt.keyCol, tt.key, gotNumRows, tt.wantNumRows)
			}
		})
	}
}

func TestString(t *testing.T) {
	conf := openTestFile(t, "sample.conf")

	row, err := conf.LookupFirst(hba.Database, "replication")
	if err != nil {
		t.Fatalf(`LookupFirst(hba.Database, "replication") errored with '%s', wanted no error`, err)
	}

	tests := []struct {
		name    string
		col     int
		want    string
		noerror bool
	}{
		{"Get ConnType", hba.ConnType, "host", true},
		{"Get User", hba.User, "postgres", true},
		{"Get Address", hba.Address, "127.0.0.1/32", true},
		{"Get Method", hba.Method, "md5", true},
		{"Get not existing column", 999, "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := conf.String(row, tt.col)
			if err != nil && tt.noerror {
				t.Errorf("String(row, %d) errored with '%s', wanted no error", tt.col, err)
			} else if err == nil && !tt.noerror {
				t.Errorf("String(row, %d) did not error, wanted error", tt.col)
			} else if got != tt.want {
				t.Errorf("String(row, %d) = %q, want %q", tt.col, got, tt.want)
			}
		})
	}
}

func TestAppendEntry(t *testing.T) {
	conf := openTestFile(t, "sample.conf")

	_, err := conf.AppendEntry("hostssl", "replication", "replication", "10.0.0.4/32", "md5")
	if err != nil {
		t.Fatalf("AppendEntry() errored with '%s', wanted no error", err)
	}

	got := conf.All()
	appended, err := regexp.MatchString(`hostssl\sreplication\sreplication\s10.0.0.4/32\smd5`, got)
	if !appended || err != nil {
		t.Errorf("AppendEntry() = failed to append a new row")
	}
}
