package strparam

import (
	"bytes"
	"fmt"
)

// TokenSchemaString returns string of token (excluding parameter values).
//
// Can be used as a key for a hash from a token schema
func TokenSchemaString(t Token) string {
	switch t.Mode {
	case UNKNOWN_TokenMode:
		return ""
	case CONST:
		return fmt.Sprintf("Const(%q, len=%d)", t.Raw, t.Len)
	case SEPARATOR:
		return fmt.Sprintf("Separator(%q, len=%d)", t.Raw, t.Len)
	case PARAMETER:
		return fmt.Sprintf("Param(%q)", t.ParamName())
	case PARAMETER_PARSED:
		// this is primarily a parameter
		return fmt.Sprintf("Param(%q)", t.ParamName())
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

// ListTokensSchemaString returns string of list tokens (excluding parameter values).
//
// Can be used as a key for a hash from a token schema
func ListTokensSchemaString(t Tokens) string {
	res := new(bytes.Buffer)
	for i, token := range t {
		if i > 0 {
			fmt.Fprint(res, "->")
		}
		fmt.Fprint(res, TokenSchemaString(token))
	}
	return res.String()
}
