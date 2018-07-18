package pgconf

import "fmt"

// ErrKeyNotFound is returned if the key being looked up cannot be found in the configuration file.
var ErrKeyNotFound = fmt.Errorf("key not found")

// ErrKeyWithoutValue is returned if the key being looked up is found, but has no value.
// Keys with no values are not supported and the line containing it is ignored.
var ErrKeyWithoutValue = fmt.Errorf("key without value")

// ErrEmptyLine is returned if a line contains no key (eg. it is empty or contains only a comment/whitespace).
var ErrEmptyLine = fmt.Errorf("no key found")
