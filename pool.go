package strparam

import "sync"

var MaxListParamsCap = 32
var MaxListTokensCap = 128

var (
	listParamsPool = sync.Pool{
		New: func() interface{} { return make([]Param, 0, MaxListParamsCap) },
	}
	listTokensPool = sync.Pool{
		New: func() interface{} { return make([]Token, 0, MaxListTokensCap) },
	}
)

func getListParams() (v []Param) {
	ifc := listParamsPool.Get()
	if ifc != nil {
		v = ifc.([]Param)
	}
	return
}

func putListParams(v []Param) {
	if cap(v) <= MaxListParamsCap {
		v = v[:0]
		listParamsPool.Put(v)
	}
}

func getlistTokens() (v []Token) {
	ifc := listTokensPool.Get()
	if ifc != nil {
		v = ifc.([]Token)
	}
	return
}

func putlistTokens(v []Token) {
	if cap(v) <= MaxListTokensCap {
		v = v[:0]
		listTokensPool.Put(v)
	}
}

func (s *Store) getlistTokens() (v []Token) {
	ifc := s.tokensPool.Get()
	if ifc != nil {
		v = ifc.([]Token)
	}
	return
}

func (s *Store) putlistTokens(v []Token) {
	if cap(v) <= MaxListTokensCap {
		v = v[:0]
		s.tokensPool.Put(v)
	}
}
