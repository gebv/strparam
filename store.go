package strparam

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/pkg/errors"
)

func NewStore() *Store {
	return &Store{
		root:       &node{Token: Token{}},
		tokensPool: sync.Pool{},
	}
}

func (r *Store) Add(in string) error {
	schema, err := Parse(in)
	if err != nil {
		return errors.Wrap(err, "failed parse")
	}

	if len(schema.Tokens) > r.maxSize {
		r.maxSize = len(schema.Tokens)
		r.tokensPool.New = func() interface{} { return make([]Token, 0, r.maxSize) }
	}

	appendChild(r.root, 0, schema.Tokens)

	return nil
}

// func (r *Store) Handler(in string, fn func(params Params)) error {
// 	schema, err := r.Find(in)
// 	if err != nil {
// 		return err
// 	}
// 	ok, params := schema.Lookup(in)
// 	if ok {
// 		fn(params)
// 	}
// 	return nil
// }

func (r *Store) Find(in string) (*PatternSchema, error) {
	tokens := r.getlistTokens()
	numParams := 0

	lookupNextToken(in, 0, r.root, &tokens, &numParams)
	defer r.putlistTokens(tokens)

	if len(tokens) > 0 {
		tokens = append(Tokens{StartEndTokens[0]}, append(tokens, StartEndTokens[1])...)
	}

	return &PatternSchema{Tokens: tokens, NumParams: numParams}, nil
}

func lookupNextToken(in string, offset int, parent *node, res *[]Token, numParams *int) {
	for _, node := range parent.Childs {
		switch node.Token.Mode {
		case BEGINLINE:
			lookupNextToken(in, offset+node.Token.Len, node, res, numParams)

			// returns because must be onece start token
			return
		case ENDLINE:
			// returns because have reached the end
			return
		case PATTERN:
			if offset <= len(in) && offset+node.Token.Len <= len(in) {
				if in[offset:offset+node.Token.Len] == node.Token.Raw {
					*res = append(*res, node.Token)
					lookupNextToken(in, offset+node.Token.Len, node, res, numParams)
					// jump down the tree
					return
				}
			}

		case PARAMETER:
			nextPattern, paramWidth := lookupNextPattern(in, offset, node)

			if offset <= len(in) && offset+paramWidth <= len(in) {
				*res = append(*res, Token{
					Mode:  PARAMETER_PARSED,
					Len:   paramWidth,
					Raw:   in[offset : offset+paramWidth],
					Param: &node.Token,
				})
				*numParams++
			}

			if paramWidth > 0 {
				*res = append(*res, nextPattern.Token)
				lookupNextToken(in, offset+paramWidth+nextPattern.Token.Len, nextPattern, res, numParams)
			} else {
				lookupNextToken(in, offset, node, res, numParams)
			}

			return
		default:
			panic(fmt.Sprintf("not supported token type %v", node.Token.Mode))
		}
	}
}

func lookupNextPattern(in string, offset int, param *node) (*node, int) {
	for _, node := range param.Childs {
		switch node.Token.Mode {
		case BEGINLINE:
			panic("that is impossible: beginning of line in the middle of a word")
		case ENDLINE:
			if param.Token.Mode == PARAMETER {
				// tail is the parameter value - because parameter is the last in the pattern
				return node, len(in) - offset
			}
			return node, 0
		case PARAMETER:
			panic("out of sequence parameter")
		case PATTERN:
			if offset > len(in) {
				return node, 0
			}
			if found := strings.Index(in[offset:], node.Token.Raw); found > -1 {
				return node, found
			}
		}
	}
	return nil, 0
}

// TODO: cover with tests as the tree is filled
func appendChild(parent *node, i int, tokens []Token) {
	if i >= len(tokens) {
		return
	}

	for _, node := range parent.Childs {
		if node.Token.Equal(tokens[i]) {
			appendChild(node, i+1, tokens)
			return
		}
	}

	newNode := &node{Token: tokens[i]}
	parent.Childs = append(parent.Childs, newNode)
	appendChild(newNode, i+1, tokens)
}

type Store struct {
	root *node
	// max size slice of tokens for all patterns
	maxSize    int
	tokensPool sync.Pool
}

func (s *Store) String() string {
	res := new(bytes.Buffer)
	dumpChilds(res, 0, s.root)
	return res.String()
}

func dumpChilds(w io.Writer, level int, n *node) {
	fmt.Fprintln(w, strings.Repeat("\t", level), n.Token.String())
	for _, child := range n.Childs {
		dumpChilds(w, level+1, child)
	}
}

type node struct {
	Token  Token
	Childs []*node
}
