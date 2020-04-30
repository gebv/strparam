package strparam

import (
	"strings"

	"github.com/pkg/errors"
)

func NewStore() *Store {
	return &Store{
		root: &node{Token: Token{}},
	}
}

func (r *Store) Add(in string) error {
	schema, err := Parse(in)
	if err != nil {
		return errors.Wrap(err, "failed parse")
	}

	appendChild(r.root, 0, schema.Tokens)

	return nil
}

func (r *Store) Find(in string) (*PatternSchema, error) {
	res := []Token{} // TODO: save maximum size
	numParams := 0

	lookupNextToken(in, 0, r.root, &res, &numParams)

	return &PatternSchema{Tokens: res, NumParams: numParams}, nil
}

func lookupNextToken(in string, offset int, parent *node, res *[]Token, numParams *int) {
	for _, node := range parent.Childs {
		switch node.Token.Mode {
		case BEGINLINE:
			lookupNextToken(in, offset+node.Token.Len, node, res, numParams)
			return
		case ENDLINE:
			return
		case PATTERN:
			if in[offset:offset+node.Token.Len] == node.Token.Raw {
				*res = append(*res, node.Token)
				lookupNextToken(in, offset+node.Token.Len, node, res, numParams)
				return
			}
		case PARAMETER:
			if nextPattern, paramWidth := lookupNextPattern(in, offset, node); paramWidth > 0 {
				*res = append(*res, Token{
					Mode:  PARAMETER_PARSED,
					Len:   paramWidth,
					Raw:   in[offset : offset+paramWidth],
					Param: &node.Token,
				})
				*numParams++
				*res = append(*res, nextPattern.Token)
				lookupNextToken(in, offset+paramWidth+nextPattern.Token.Len, nextPattern, res, numParams)
				return
			}
		}
	}
}

func lookupNextPattern(in string, offset int, param *node) (*node, int) {
	for _, node := range param.Childs {
		switch node.Token.Mode {
		case BEGINLINE:
			panic("that is impossible: beginning of line in the middle of a word")
		case ENDLINE:
			return node, 0
		case PARAMETER:
			panic("out of sequence parameter")
		case PATTERN:
			if found := strings.Index(in[offset:], node.Token.Raw); found > -1 {
				return node, found
			}
		}
	}
	return nil, 0
}

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
}

type node struct {
	Token  Token
	Childs []*node
}
