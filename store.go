package strparam

import (
	"bytes"
	"fmt"
	"io"
	"sort"
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

func (r *Store) AddPattern(p *Pattern) {
	if len(p.Tokens) > r.maxSize {
		r.maxSize = len(p.Tokens)
		r.tokensPool.New = func() interface{} { return make([]Token, 0, r.maxSize) }
	}

	appendChild(r.root, 0, p.Tokens)
}

func (r *Store) add(name, exp string) (*Pattern, error) {
	schema, err := ParseWithName(name, exp)
	if err != nil {
		return nil, errors.Wrap(err, "failed parse")
	}

	r.AddPattern(schema)

	return schema, nil
}

// Find returns full pattern matched for incoming string.
func (r *Store) Find(in string) *Pattern {
	tokens := r.getlistTokens()
	numParams := 0

	lookupNextToken(in, 0, r.root, &tokens, &numParams)
	defer r.putlistTokens(tokens)

	if len(tokens) <= 2 || tokens[0].Mode != START || tokens[len(tokens)-1].Mode != END {
		// not a complete pattern
		return nil
	}

	return &Pattern{Tokens: tokens, NumParams: numParams}
}

func lookupNextToken(in string, offset int, parent *node, res *[]Token, numParams *int) {
	// if offset >= len(in) {
	// 	log.Printf("Offset %d has gone out of bounds (or is equal) of the incoming string (len=%d).\n", offset, len(in))
	// 	return
	// }

	for idx, child := range parent.Childs {
		switch child.Token.Mode {
		case START:
			// general case
			//
			// -- {START}
			// -- -- {CONST}
			// -- -- -- {CONST}
			// -- -- -- -- {END}
			// -- -- -- {END}
			// -- -- {CONST}
			// -- -- -- {PARAM}
			// -- -- -- -- {END}
			// -- -- -- {END}
			// -- -- {PARAM}
			// -- -- -- {CONST}
			// -- -- -- -- {END}
			// -- -- -- {END}
			// -- -- {END}

			*res = append(*res, child.Token)

			// jump into the branch
			lookupNextToken(in, offset, child, res, numParams)

			// returns because must be onece start token
			return
		case END:
			// general case same as for START, but analize from END
			// only if the offset is strictly equal to the input string
			if len(in) == offset {
				// if we have reached the END type token, then we have completely specific pattern
				*res = append(*res, child.Token)
				// returns because have reached the end
				return
			}
		case CONST, SEPARATOR:
			// general case
			//
			// -- {CONST} <- look here
			// -- -- {PARAM}
			// -- -- -- {CONST}
			// -- -- -- {END}
			// -- -- {CONST}
			// -- -- {END}

			if offset+child.Token.Len <= len(in) {
				if in[offset:offset+child.Token.Len] == child.Token.Raw {
					*res = append(*res, child.Token)

					// jump into the branch
					lookupNextToken(in, offset+child.Token.Len, child, res, numParams)

					// returns because we move deeper into the tree
					return
				}
			}

		case PARAMETER:
			// general case
			//
			// -- {PARAM} <- look here
			// -- -- {CONST}
			// -- -- {END}

			// looking for the next node to understand when the parameter ends
			nextNode, addOffset := rightPath(in, offset, child)

			if nextNode != nil && len(in) >= offset+addOffset {
				// if offset+addOffset+nextNode.Token.Len > len(in) {
				// 	continue
				// }

				if len(parent.Childs)-1 > idx && addOffset == 0 {
					continue
				}

				*res = append(*res, Token{
					Mode:  PARAMETER_PARSED,
					Len:   addOffset,
					Raw:   in[offset : offset+addOffset],
					Param: &child.Token,
				})
				*numParams++

				// added const or END token (that after the parameter)
				*res = append(*res, nextNode.Token)

				// jump to found const token
				lookupNextToken(in, offset+addOffset+nextNode.Token.Len, nextNode, res, numParams)

				// returns because we move deeper into the tree from found matched pattern
				return
			}

			// // we not found next pattern,
			// lookupNextToken(in, offset, child, res, numParams)

			// // returns because we move deeper into the tree from current node
			// return
		default:
			panic(fmt.Sprintf("not supported token type %v", child.Token.Mode))
		}
	}
}

// func lookupNextToken(in string, offset int, parent *node, res *[]Token, numParams *int) {
// 	log.Println(">>>> lookupNextToken", parent.Token.String())
// 	for _, node := range parent.Childs {
// 		log.Println("\t>>>>>", parent.Token.String())
// 		log.Println("\t>>>>>", node.Token.String())
// 		log.Println("\t>>>>>", len(in), offset, offset+node.Token.Len)

// 		switch node.Token.Mode {
// 		case START:
// 			*res = append(*res, node.Token)

// 			// jump into the branch
// 			lookupNextToken(in, offset+node.Token.Len, node, res, numParams)

// 			// returns because must be onece start token
// 			return
// 		case END:
// 			// only if the offset is strictly equal to the input string
// 			if len(in) == offset {
// 				// if we have reached the END type token, then we have completely specific pattern
// 				*res = append(*res, node.Token)
// 				// returns because have reached the end
// 				return
// 			}
// 		case CONST:
// 			// general case
// 			//
// 			// -- {CONST}
// 			// -- -- {PARAM}
// 			// -- -- {END}
// 			if offset <= len(in) && offset+node.Token.Len <= len(in) {
// 				if in[offset:offset+node.Token.Len] == node.Token.Raw {
// 					if node.isOneEndChild() && len(in) != offset+node.Token.Len {
// 						// go to next child for current level if the branch ended and not matched lengths for cursor and input value
// 						continue
// 					}

// 					*res = append(*res, node.Token)

// 					// jump into the branch
// 					lookupNextToken(in, offset+node.Token.Len, node, res, numParams)

// 					// returns because we move deeper into the tree
// 					return
// 				}
// 			}

// 		case PARAMETER:
// 			// general case
// 			//
// 			// -- {PARAM}
// 			// -- -- {CONST}
// 			// -- -- {END}
// 			nextPattern, paramWidth := lookupNextPattern(in, offset, node)

// 			if offset <= len(in) && offset+paramWidth <= len(in) {
// 				*res = append(*res, Token{
// 					Mode:  PARAMETER_PARSED,
// 					Len:   paramWidth,
// 					Raw:   in[offset : offset+paramWidth],
// 					Param: &node.Token,
// 				})
// 				*numParams++
// 			}

// 			if paramWidth > 0 {
// 				*res = append(*res, nextPattern.Token)
// 				lookupNextToken(in, offset+paramWidth+nextPattern.Token.Len, nextPattern, res, numParams)
// 				// returns because we move deeper into the tree from found matched pattern
// 				return
// 			} else {
// 				lookupNextToken(in, offset, node, res, numParams)
// 				// returns because we move deeper into the tree from current node
// 				return
// 			}
// 		default:
// 			panic(fmt.Sprintf("not supported token type %v", node.Token.Mode))
// 		}
// 	}
// }

// func lookupNextPattern(in string, offset int, param *node) (*node, int) {
// 	for _, node := range param.Childs {
// 		switch node.Token.Mode {
// 		case START:
// 			panic("that is impossible: beginning of line in the middle of a word")
// 		case END:
// 			// TODO: если встречаем `/` то обрубаем
// 			if param.Token.Mode == PARAMETER {
// 				// tail is the parameter value - because parameter is the last in the pattern
// 				return node, len(in) - offset
// 			}
// 			return node, 0
// 		case PARAMETER:
// 			panic("out of sequence parameter")
// 		case CONST:
// 			if offset > len(in) {
// 				// out of bounds
// 				return node, 0
// 			}
// 			// TODO: если встречаем `/` то прерываем
// 			if found := strings.Index(in[offset:], node.Token.Raw); found > -1 {
// 				return node, found
// 			}
// 		}
// 	}
// 	return nil, 0
// }

func rightPath(in string, offset int, node *node) (*node, int) {
	for _, child := range node.Childs {
		switch child.Token.Mode {
		// case PARAMETER:
		// 	// -- {CONST}
		// 	// -- -- {PARAM} <- look here
		// 	// -- -- -- {CONST}
		// 	// -- -- -- -- {END}
		// 	// -- -- -- {END}

		// 	nextNode, addOffset := rightPath(in, offset, child)
		// 	if nextNode != nil && len(in) >= offset+addOffset {
		// 		switch nextNode.Token.Mode {
		// 		case END:
		// 			return child, 0
		// 		case CONST:
		// 			if found := strings.Index(in[offset:], child.Token.Raw); found > -1 {
		// 				return child, 0
		// 			}
		// 		}
		// 	}

		case CONST, SEPARATOR:
			// -- {CONST} <- look here
			// -- -- {PARAM}
			// -- -- {CONST}
			// -- -- {END}
			if found := strings.Index(in[offset:], child.Token.Raw); found > -1 {
				return child, found
			}
		case END:
			// returns tail
			return child, len(in) - offset
		default:
			panic(fmt.Errorf("not expected node type %q", child.Token.Mode.String()))
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

	sort.Sort(parent)
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
	printChilds(res, 0, s.root)
	return res.String()
}

// helper function to write node and childs of current branch.
func printChilds(w io.Writer, level int, n *node) {
	fmt.Fprintln(w, strings.Repeat("\t", level), n.Token.String())
	for _, child := range n.Childs {
		printChilds(w, level+1, child)
	}
}

// tree node
type node struct {
	Token  Token
	Childs []*node
}

// // isOneEndChild reutrns true if the current branch has END
// func (n *node) isOneEndChild() bool {
// 	if len(n.Childs) == 1 {
// 		return n.Childs[0].Token.Mode == END
// 	}
// 	return false
// }

func (n *node) Len() int {
	return len(n.Childs)
}

func (n *node) lengthConstOrZero() int {
	if n.Token.Mode != CONST {
		return 0
	}
	return len(n.Token.Raw)
}

func (n *node) Less(i, j int) bool {
	if n.Childs[i].Token.Mode != n.Childs[j].Token.Mode {
		if n.Childs[i].Token.Mode == CONST {
			return true
		}
	}
	if n.Childs[i].lengthConstOrZero() != n.Childs[j].lengthConstOrZero() {
		return n.Childs[i].lengthConstOrZero() >= n.Childs[j].lengthConstOrZero()
	}
	return len(n.Childs[i].Childs) >= len(n.Childs[j].Childs)
}

func (n *node) Swap(i, j int) {
	n.Childs[i], n.Childs[j] = n.Childs[j], n.Childs[i]
}

var _ sort.Interface = (*node)(nil)
