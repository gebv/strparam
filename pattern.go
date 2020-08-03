package strparam

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"
)

// ParseWithName analyzes the pattern, split it into tokens. Saves the schema name.
//
// Iterate over the UTF-8 characters (with correct offset of bytes).
func ParseWithName(name, exp string) (*Pattern, error) {
	return parse(name, exp)
}

// Parse analyzes the pattern, split it into tokens.
//
// Iterate over the UTF-8 characters (with correct offset of bytes).
func Parse(exp string) (*Pattern, error) {
	return parse("", exp)
}

func parse(patternName, exp string) (*Pattern, error) {
	if exp == "" {
		return nil, errors.New("expression should not is empty")
	}

	tokens := getlistTokens()
	defer putlistTokens(tokens)

	// end and start of parameter positions in bytes
	var start, end int = 0, 0
	// current mode (initial as Pattern)
	var mode TokenMode = CONST
	// is flag of start char of input string
	var EOF bool
	// number of parameters
	var numParams int

	// start of input string
	tokens = append(tokens, Token{
		Mode: START,
	})

	// current UTF-8 character in a word
	var char rune

	// w - character width in bytes
	for i, w := 0, 0; i < len(exp); i += w {
		char, w = utf8.DecodeRuneInString(exp[i:])
		EOF = i == len(exp)-1

		switch char {
		case DefaultStartParam:

			// invalid input string if after end border of parameter got new parameter
			if i > 0 && i-end == 0 {
				return nil, fmt.Errorf("should be a pattern between the parameters, pos %d", i)
			}

			if i-end > 0 {
				// skip empty pattern
				// - at beginning of the input string the parameter
				tokens = append(tokens, Token{
					Mode: CONST,
					Len:  i - end,
					Raw:  exp[end:i],
				})
			}

			mode = PARAMETER
			start = i // sets start position of parameter
		case DefaultEndParam:
			if mode != PARAMETER {
				// skip single closing characters of parameters
				continue
			}

			// empty name of parameter if after start border of parameter got end border
			if start-i+1 == 0 {
				return nil, fmt.Errorf("empty name of parameter, pos %d", i)
			}

			tokens = append(tokens, Token{
				Mode: PARAMETER,
				Raw:  exp[start : i+1],
			})

			mode = CONST
			end, start = i+1, i+1 // zeroing positions
			numParams++
		}

		if EOF {
			// invalid parameter if EOF before closed parameter
			if mode == PARAMETER {
				return nil, fmt.Errorf("parameter was not closed, pos %d", i)
			}

			if mode == CONST {
				// if exists chars after closed parameter
				if i+1-end > 0 {
					tokens = append(tokens, Token{
						Mode: CONST,
						Len:  i + 1 - end,
						Raw:  exp[end : i+1],
					})
				}
			}
		}
	}

	// end of input string
	tokens = append(tokens, Token{
		Mode: END,
		Raw:  patternName,
	})

	return &Pattern{
		Tokens:    tokens,
		NumParams: numParams,
	}, nil
}

// Lookup returns list params if input string matched to schema.
//
// NOTE: nothing (empty list of tokens) not matches to anything.
func (s *Pattern) Lookup(in string) (bool, Params) {
	if s == nil {
		return false, nil
	}

	params := getListParams()
	defer putListParams(params)

	if len(s.Tokens) == 0 {
		// nothing not matches to anything
		return false, nil
	}

	// this is the sum of the lengths of the patterns and found value of parameters
	var offset int

	for num, t := range s.Tokens {
		if len(in) < offset || len(in) < offset+t.Len {
			return false, nil
		}

		switch t.Mode {
		case START:
		case END:
			goto exitloop
		case PARAMETER_PARSED:
			params = append(params, Param{
				Name:  t.ParamName(),
				Value: in[offset : offset+t.Len],
			})
			offset += t.Len
		case PARAMETER:
			_next := s.Tokens[num+1]
			switch _next.Mode {
			case END:
				params = append(params, Param{
					Name:  t.ParamName(),
					Value: in[offset:],
				})
				offset += len(in) - offset
			case CONST, SEPARATOR:
				if found := strings.Index(in[offset:], _next.Raw); found > -1 {
					params = append(params, Param{
						Name:  t.ParamName(),
						Value: in[offset : offset+found],
					})
					// add the length of the found parameter value
					offset += found
				}
			case PARAMETER:
				panic("should be a pattern between the parameters")
			default:
				return false, nil
			}
		case CONST, SEPARATOR:
			if in[offset:offset+t.Len] == t.Raw {
				// add the length of the pattern
				offset += t.Len
			} else {
				// forced return because pattern is not matched
				return false, nil
			}
		}
	}

exitloop:

	// offset did not seeking to end of by input value
	if len(in) != offset {
		return false, nil
	}

	// received an unexpected number of parameters
	if len(params) != s.NumParams {
		return false, nil
	}

	return true, params
}

type Pattern struct {
	Tokens    Tokens
	NumParams int
}

// String returns schema of pattern.
func (s Pattern) String() string {
	return s.Tokens.String()
}

// Name returns name of pattern (by end token if sets).
func (s Pattern) Name() string {
	if s.Tokens == nil {
		return ""
	}
	if len(s.Tokens) <= 1 {
		return ""
	}
	endToken := s.Tokens[len(s.Tokens)-1]
	if endToken.Mode != END {
		return ""
	}
	return endToken.Raw
}

type Param struct {
	Name  string
	Value string
}

type Params []Param
