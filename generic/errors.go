package generic

import "fmt"

// ErrKeyNotFound is returned if the key being looked up cannot be found in the configuration file.
var ErrKeyNotFound = fmt.Errorf("key not found")

// ErrEmptyLine is returned if a line contains no key (eg. it is empty or contains only a comment/whitespace).
var ErrEmptyLine = fmt.Errorf("no key found")
