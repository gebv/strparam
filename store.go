package strparam

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/pkg/errors"
)

// NewStore returns new storage instanse for patterns.
func NewStore() *Store {
	return &Store{
		root:       &node{Token: Token{}},
		tokensPool: sync.Pool{},
	}
}

// Add returns parsed and added pattern from input value.
//
// Error is returned if parsing error.
func (r *Store) Add(exp string) (*Pattern, error) {
	return r.add("", exp)
}

// AddNamed add named new pattern.
func (r *Store) AddNamed(name, exp string) (*Pattern, error) {
	return r.add(name, exp)
}

func (r *Store) add(name, exp string) (*Pattern, error) {
	schema, err := ParseWithName(name, exp)
	if err != nil {
		return nil, errors.Wrap(err, "failed parse")
	}

	if len(schema.Tokens) > r.maxSize {
		r.maxSize = len(schema.Tokens)
		r.tokensPool.New = func() interface{} { return make([]Token, 0, r.maxSize) }
	}

	appendChild(r.root, 0, schema.Tokens)

	return schema, nil
}

// Find returns full pattern matched for incoming string.
func (r *Store) Find(in string) *Pattern {
	tokens := r.getlistTokens()
	numParams := 0

	lookupNextToken(in, 0, r.root, &tokens, &numParams)
	defer r.putlistTokens(tokens)

	if len(tokens) <= 2 || tokens[0].Mode != START || tokens[len(tokens)-1].Mode != END {
		return nil
	}

	return &Pattern{Tokens: tokens, NumParams: numParams}
}

func lookupNextToken(in string, offset int, parent *node, res *[]Token, numParams *int) {
	for _, node := range parent.Childs {
		switch node.Token.Mode {
		case START:
			*res = append(*res, node.Token)

			// jump into the branch
			lookupNextToken(in, offset+node.Token.Len, node, res, numParams)

			// returns because must be onece start token
			return
		case END:
			// only if the offset is strictly equal to the input string
			if len(in) == offset {
				// if we have reached the END type token, then we have completely specific pattern
				*res = append(*res, node.Token)
				// returns because have reached the end
				return
			}
		case CONST:
			if offset <= len(in) && offset+node.Token.Len <= len(in) {
				if in[offset:offset+node.Token.Len] == node.Token.Raw {
					if node.isOneEndChild() && len(in) != offset+node.Token.Len {
						// go to next child for current level if the branch ended and not matched lengths for cursor and input value
						continue
					}

					*res = append(*res, node.Token)

					// jump into the branch
					lookupNextToken(in, offset+node.Token.Len, node, res, numParams)

					// returns because we move deeper into the tree
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
				// returns because we move deeper into the tree from found matched pattern
				return
			} else {
				lookupNextToken(in, offset, node, res, numParams)
				// returns because we move deeper into the tree from current node
				return
			}
		default:
			panic(fmt.Sprintf("not supported token type %v", node.Token.Mode))
		}
	}
}

func lookupNextPattern(in string, offset int, param *node) (*node, int) {
	for _, node := range param.Childs {
		switch node.Token.Mode {
		case START:
			panic("that is impossible: beginning of line in the middle of a word")
		case END:
			if param.Token.Mode == PARAMETER {
				// tail is the parameter value - because parameter is the last in the pattern
				return node, len(in) - offset
			}
			return node, 0
		case PARAMETER:
			panic("out of sequence parameter")
		case CONST:
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

// Store this is patterns repository.
type Store struct {
	root *node
	// max size slice of tokens for all patterns
	maxSize    int
	tokensPool sync.Pool
}

// String returns the patent storage schema as a tree.
func (s *Store) String() string {
	res := new(bytes.Buffer)
	dumpChilds(res, 0, s.root)
	return res.String()
}

// helper function to write node and childs of current branch.
func dumpChilds(w io.Writer, level int, n *node) {
	fmt.Fprintln(w, strings.Repeat("\t", level), n.Token.String())
	for _, child := range n.Childs {
		dumpChilds(w, level+1, child)
	}
}

// tree node
type node struct {
	Token  Token
	Childs []*node
}

// isOneEndChild reutrns true if the current branch has END
func (n *node) isOneEndChild() bool {
	if len(n.Childs) == 1 {
		return n.Childs[0].Token.Mode == END
	}
	return false
}
