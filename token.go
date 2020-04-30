package strparam

import "fmt"

var DefaultStartParam = '{'
var DefaultEndParam = '}'

type Token struct {
	Mode TokenMode
	// len of bytes
	Len   int
	Raw   string
	Param *Token
}

func (t Token) Equal(in Token) bool {
	return t.Mode == in.Mode && t.Len == in.Len && t.Raw == in.Raw
}

func (t *Token) String() string {
	switch t.Mode {
	case UNKNOWN_TokenMode:
		return ""
	case PATTERN:
		return fmt.Sprintf("Pattern(%q, len=%d)", t.Raw, t.Len)
	case PARAMETER:
		return fmt.Sprintf("Parameter(%q)", t.Raw)
	case BEGINLINE:
		return fmt.Sprintf("START")
	case ENDLINE:
		return fmt.Sprintf("END")
	}

	return fmt.Sprintf("%#v", t)
}

// ParamName returns parameter name if mode of current token is PARAMETER.
// In all other cases returns an empty value.
func (t *Token) ParamName() string {
	if t.Mode == PARAMETER && len(t.Raw) > 2 {
		return t.Raw[1 : len(t.Raw)-1]
	}
	if t.Mode == PARAMETER_PARSED && len(t.Param.Raw) > 2 {
		return t.Param.Raw[1 : len(t.Param.Raw)-1]
	}
	return ""
}

type TokenMode int

func (m TokenMode) String() string {
	switch m {
	case PATTERN:
		return "pattern"
	case PARAMETER:
		return "paremeter"
	case BEGINLINE:
		return "begin"
	case ENDLINE:
		return "end"
	}

	return fmt.Sprintf("TokenMode(%d)", m)
}

const (
	UNKNOWN_TokenMode TokenMode = 0
	// PATTERN boarder of pattern
	PATTERN TokenMode = 1
	// PARAMETER boarder of parameter
	PARAMETER TokenMode = 2
	// BEGINLINE marker of begin line
	BEGINLINE TokenMode = 4
	// ENDLINE marker of end line
	ENDLINE TokenMode = 5
	// PARAMETER_PARSED with known offsets
	PARAMETER_PARSED TokenMode = 6
)
