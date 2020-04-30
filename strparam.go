package strparam

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// Parse analyzes the pattern, split it into tokens.
//
// Iterate over the UTF-8 characters (with correct offset of bytes).
func Parse(in string) (*PatternSchema, error) {
	tokens := getlistTokens()
	defer putlistTokens(tokens)

	// end and start of parameter positions in bytes
	var start, end int = 0, 0
	// current mode
	var mode TokenMode
	// is flag of start char of input string
	var EOF bool
	// number of parameters
	var numParams int

	// end of input string
	tokens = append(tokens, Token{
		Mode: BEGINLINE,
	})

	// current UTF-8 character in a word
	var char rune

	// w - character width in bytes
	for i, w := 0, 0; i < len(in); i += w {
		char, w = utf8.DecodeRuneInString(in[i:])
		EOF = i == len(in)-1

		switch char {
		case DefaultStartParam:

			// invalid input string if after end border of parameter got new parameter
			if i-end == 0 {
				return nil, fmt.Errorf("should be a pattern between the parameters, pos %d", i)
			}

			tokens = append(tokens, Token{
				Mode: PATTERN,
				Len:  i - end,
				Raw:  in[end:i],
			})

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
				Raw:  in[start : i+1],
			})

			mode = PATTERN
			end, start = i+1, i+1 // zeroing positions
			numParams++
		}

		if EOF {
			// invalid parameter if EOF before closed parameter
			if mode == PARAMETER {
				return nil, fmt.Errorf("parameter was not closed, pos %d", i)
			}

			if mode == PATTERN {
				// if exists chars after closed parameter
				if i+1-end > 0 {
					tokens = append(tokens, Token{
						Mode: PATTERN,
						Len:  i + 1 - end,
						Raw:  in[end : i+1],
					})
				}
			}
		}
	}

	// start of input string
	tokens = append(tokens, Token{
		Mode: ENDLINE,
	})

	return &PatternSchema{
		Pattern:   in,
		Tokens:    tokens,
		NumParams: numParams,
	}, nil
}

// Lookup returns list params if input string matched to schema.
func (s *PatternSchema) Lookup(in string) (bool, Params) {
	params := getListParams()
	defer putListParams(params)

	// this is the sum of the lengths of the patterns and found value of parameters
	var offset int

	for num, t := range s.Tokens {
		switch t.Mode {
		case BEGINLINE:
		case ENDLINE:
			if offset < len(in) {
				// have tail without pattern
				return false, nil
			}
		case PARAMETER_PARSED:
			params = append(params, Param{
				Name:  t.ParamName(),
				Value: in[offset : offset+t.Len],
			})
			offset += t.Len
		case PARAMETER:
			_next := s.Tokens[num+1]
			switch _next.Mode {
			case ENDLINE:
				params = append(params, Param{
					Name:  t.ParamName(),
					Value: in[offset:],
				})
				offset += len(in) - offset
			case PATTERN:
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
		case PATTERN:
			if in[offset:offset+t.Len] == t.Raw {
				// add the length of the pattern
				offset += t.Len
			} else {
				return false, nil
			}
		}
	}

	if len(params) != s.NumParams {
		return false, nil
	}
	return true, params
}

type PatternSchema struct {
	Pattern   string
	Tokens    []Token
	NumParams int
}

type Param struct {
	Name  string
	Value string
}

type Params []Param
