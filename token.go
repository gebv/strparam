package strparam

import (
	"bytes"
	"fmt"
)

var DefaultStartParam = '{'
var DefaultEndParam = '}'

type Token struct {
	Mode TokenMode
	// len of bytes
	Len int
	// multifunctional field
	Raw   string
	Param *Token
}

// Equal returns true if mode and length and values is equal.
func (t Token) Equal(in Token) bool {
	return t.Mode == in.Mode && t.Len == in.Len && t.Raw == in.Raw
}

// String returns human-readable format of token .
func (t *Token) String() string {
	switch t.Mode {
	case UNKNOWN_TokenMode:
		return ""
	case CONST:
		return fmt.Sprintf("Const(%q, len=%d)", t.Raw, t.Len)
	case SEPARATOR:
		return fmt.Sprintf("Separator(%q, len=%d)", t.Raw, t.Len)
	case PARAMETER:
		return fmt.Sprintf("Param(%q)", t.Raw)
	case PARAMETER_PARSED:
		return fmt.Sprintf("ParsedParam(%s=%q)", t.ParamName(), t.Raw)
	case START:
		return fmt.Sprintf("START")
	case END:
		if t.Raw != "" {
			// named token
			return fmt.Sprintf("END(%q)", t.Raw)
		}
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

// Tokens helper type for list tokens.
type Tokens []Token

func (t Tokens) String() string {
	res := new(bytes.Buffer)
	for i, token := range t {
		if i > 0 {
			fmt.Fprint(res, "->")
		}
		fmt.Fprint(res, token.String())
	}
	return res.String()
}

var (
	StartToken  = Token{Mode: START}
	EndToken    = Token{Mode: END}
	EmptySchema = Tokens{StartToken, EndToken}
)

// TokenMode type of token mode.
type TokenMode int

// String returns human-readable format of token mode.
func (m TokenMode) String() string {
	switch m {
	case SEPARATOR:
		return "separator"
	case CONST:
		return "const"
	case UNKNOWN_TokenMode:
		return "empty_token?"
	case PARAMETER:
		return "param"
	case START:
		return "begin"
	case END:
		return "end"
	case PARAMETER_PARSED:
		return "parsed_param"
	}

	return fmt.Sprintf("TokenMode(%d)", m)
}

const (
	UNKNOWN_TokenMode TokenMode = 0
	// CONST boarder of pattern
	CONST TokenMode = 1
	// PARAMETER boarder of parameter
	PARAMETER TokenMode = 2
	// START marker of begin line
	START TokenMode = 4
	// END marker of end line
	END TokenMode = 5
	// PARAMETER_PARSED with known offsets
	PARAMETER_PARSED TokenMode = 6
	// SEPARATOR special token for separating constants
	SEPARATOR TokenMode = 7
)

// ConstToken returns a token of type CONST.
func ConstToken(in string) Token {
	return Token{
		Mode: CONST,
		Len:  len(in),
		Raw:  in,
	}
}

// SeparatorToken returns a token of type SEPARATOR.
func SeparatorToken(in string) Token {
	return Token{
		Mode: SEPARATOR,
		Len:  len(in),
		Raw:  in,
	}
}

// ParameterToken returns a token of type PARAM.
func ParameterToken(rawName string) Token {
	return Token{
		Mode: PARAMETER,
		Raw:  string(DefaultStartParam) + rawName + string(DefaultEndParam),
	}
}

// ParsedParameterToken returns a token of type PARSED_PARAM.
func ParsedParameterToken(rawName, val string) Token {
	return Token{
		Mode: PARAMETER_PARSED,
		Raw:  val,
		Len:  len(val),
		Param: &Token{
			Mode: PARAMETER,
			Raw:  string(DefaultStartParam) + rawName + string(DefaultEndParam),
		},
	}
}

// NamedEndToken returns a named token of type END.
func NamedEndToken(name string) Token {
	return Token{
		Mode: END,
		Raw:  name,
	}
}
