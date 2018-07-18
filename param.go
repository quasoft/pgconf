package pgconf

// token stores the starting and ending positions of a pices of text.
// start is inclusive, while end is exclusive.
type token struct {
	start int64
	end   int64
}

// param stores the starting and ending positions of the:
// * key name - relative to the start of the file
// * value - relative to the start of the file
type param struct {
	key   token
	value token
}

// newParam creates and initializes an empty param structure
func newParam() *param {
	p := &param{
		key: token{
			start: -1,
			end:   -1,
		},
		value: token{
			start: -1,
			end:   -1,
		},
	}
	return p
}

// valueSize returns the byte size of the value. Returns 0 if value size is unknown.
func (p *param) valueSize() int64 {
	if p.value.start == -1 || p.value.end == -1 || p.value.end <= p.value.start {
		return 0
	}
	return p.value.end - p.value.start
}
