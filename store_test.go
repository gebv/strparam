package strparam

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_StoreSingle_BasicTests(t *testing.T) {
	// result should be the same as in the case Parse and Loockup for same pattern

	tests := patternBasicCases
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s: %q->%q", tt.name, tt.pattern, tt.in), func(t *testing.T) {
			if tt.wantErr {
				t.Skip("because want error")
			}

			t.Logf("[INFO] pattern = %q, input = %q", tt.pattern, tt.in)

			s := NewStore()
			s.Add(tt.pattern)

			t.Log("[INFO] store schema:", s.String())

			schema := s.Find(tt.in)

			if schema != nil {
				t.Log("[INFO] found schema:", schema.String())
			}

			found, params := schema.Lookup(tt.in)
			if found != tt.found {
				t.Errorf("Loockup().found = %v, want %v", found, tt.found)
			}
			if !reflect.DeepEqual(params, tt.want) {
				t.Errorf("Loockup().params = %v, want %v", params, tt.want)
			}
		})
	}
}

func Test_StoreSingle_ExtendsBasicTests(t *testing.T) {
	tests := []struct {
		pattern     string
		in          string
		shouldFound bool
		wantTokens  Tokens
	}{
		{"", "", false, nil},
		{"", "123", false, nil},
		{"123", "", false, nil},
		{"foobar{param}", "", false, nil},
		{"foobar{param}", "123foobar", false, nil},
		{"{param}foobar", "foobar123", false, nil},
		{"foo{param}bar", "", false, nil},
		{"foo{param}bar", "123foobar", false, nil},
		{"foo{param}bar", "foobar123", false, nil},
		{"{param}", "", true, Tokens{StartToken, ParsedParameterToken("param", ""), EndToken}},
		{"{param}", "123", true, Tokens{StartToken, ParsedParameterToken("param", "123"), EndToken}},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%q->%q", tt.pattern, tt.in), func(t *testing.T) {
			s := NewStore()
			s.Add(tt.pattern)
			t.Log("[INFO] storage structure", s.String())
			foundSchema := s.Find(tt.in)
			if tt.shouldFound {
				require.NotEmpty(t, foundSchema)
				require.EqualValues(t, tt.wantTokens, foundSchema.Tokens)
			} else {
				require.Empty(t, foundSchema)
			}
		})
	}
}

func Test_StoreMultiple(t *testing.T) {
	// NOTE: for ease of reading consider url routing (but it is not usable for http routing)
	// NOTE: the behavior is independent of the order of appends
	// NOTE: uses named patterns to match exactly

	tests := []struct {
		namedPatterns [][]string
		in            string
		shouldFound   bool
		wantTokens    Tokens
	}{
		// empty list patterns
		{[][]string{}, "", false, nil},

		{[][]string{{"name1", "/"}}, "/", true, Tokens{StartToken, ConstToken("/"), NamedEndToken("name1")}},
		{[][]string{{"", "/"}}, "/", true, Tokens{StartToken, ConstToken("/"), EndToken}},

		{[][]string{
			{"", "/a"},
			{"", "/b"},
		}, "/a", true, Tokens{StartToken, ConstToken("/a"), EndToken}},
		{[][]string{
			{"", "/a"},
			{"", "/b"},
		}, "/c", false, nil},

		{[][]string{
			// correct priority
			{"index", "/"},
			{"paramed", "/{param}"},
		}, "/", true, Tokens{StartToken, ConstToken("/"), NamedEndToken("index")}},
		{[][]string{
			{"paramed", "/{param}"},
			{"index", "/"},
			// order matters
		}, "/", true, Tokens{StartToken, ConstToken("/"), NamedEndToken("index")}},
		{[][]string{
			{"index", "/"},
			{"paramed", "/{param}"},
		}, "/foo", true, Tokens{StartToken, ConstToken("/"), ParsedParameterToken("param", "foo"), NamedEndToken("paramed")}},

		{[][]string{
			{"index", "/"},
			{"sub", "/sub"},
		}, "/sub", true, Tokens{StartToken, ConstToken("/sub"), NamedEndToken("sub")}},
		{[][]string{
			{"index", "/"},
			{"sub", "/sub"},
		}, "/", true, Tokens{StartToken, ConstToken("/"), NamedEndToken("index")}},
		{[][]string{
			{"sub", "/sub"},
			{"index", "/"},
		}, "/", true, Tokens{StartToken, ConstToken("/"), NamedEndToken("index")}},

		{[][]string{
			{"index", "/"},
			{"path", "/path/"},
			{"pathParams", "/path/{params}"},
		}, "/", true, Tokens{StartToken, ConstToken("/"), NamedEndToken("index")}},
		{[][]string{
			{"path", "/path/"},
			{"index", "/"},
			{"pathParams", "/path/{params}"},
		}, "/", true, Tokens{StartToken, ConstToken("/"), NamedEndToken("index")}},
		{[][]string{
			{"path", "/path/"},
			{"pathParams", "/path/{params}"},
			{"index", "/"},
		}, "/", true, Tokens{StartToken, ConstToken("/"), NamedEndToken("index")}},

		{[][]string{
			{"index", "/"},
			{"path", "/path/"},
			{"pathParams", "/path/{params}"},
		}, "/notexists", false, nil},

		{[][]string{
			{"index", "/"},
			{"path", "/path/"},
			{"pathParams", "/path/{params}"},
		}, "/path/", true, Tokens{StartToken, ConstToken("/path/"), NamedEndToken("path")}},
		{[][]string{
			{"index", "/"},
			{"pathParams", "/path/{params}"},
			{"path", "/path/"},
			// order matters
		}, "/path/", true, Tokens{StartToken, ConstToken("/path/"), NamedEndToken("path")}},

		{[][]string{
			{"index", "/"},
			{"path", "/path/"},
			{"pathParams", "/path/{param}"},
		}, "/path/foo", true, Tokens{StartToken, ConstToken("/path/"), ParsedParameterToken("param", "foo"), NamedEndToken("pathParams")}},
		{[][]string{
			{"index", "/"},
			{"pathParams", "/path/{param}"},
			{"path", "/path/"},
		}, "/path/foo", true, Tokens{StartToken, ConstToken("/path/"), ParsedParameterToken("param", "foo"), NamedEndToken("pathParams")}},

		// https://github.com/gebv/strparam/issues/7
		{[][]string{
			{"a", "/{foobar}"},
			{"b", "/b"},
			{"c", "/ba"},
			{"d", "/baz"},
			{"e", "/ba{foobar}"},
		}, "/baz", true, Tokens{StartToken, ConstToken("/baz"), NamedEndToken("d")}},
		{[][]string{
			{"a", "/{foobar}"},
			{"b", "/b"},
			{"c", "/ba"},
			// {"d", "/baz"},
			{"e", "/ba{foobar}"},
		}, "/baz", true, Tokens{StartToken, ConstToken("/ba"), ParsedParameterToken("foobar", "z"), NamedEndToken("e")}},
		{[][]string{
			{"a", "/{foobar}"},
			{"b", "/b"},
			{"c", "/ba"},
			// {"d", "/baz"},
			// {"e", "/ba{foobar}"},
		}, "/baz", true, Tokens{StartToken, ConstToken("/"), ParsedParameterToken("foobar", "baz"), NamedEndToken("a")}},
		{[][]string{
			// {"a", "/{foobar}"},
			{"b", "/b"},
			{"c", "/ba"},
			// {"d", "/baz"},
			// {"e", "/ba{foobar}"},
		}, "/baz", false, nil},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%q->%q", tt.namedPatterns, tt.in), func(t *testing.T) {
			s := NewStore()
			for _, rawPattern := range tt.namedPatterns {
				s.AddNamed(rawPattern[0], rawPattern[1])
			}

			t.Log("[INFO] storage structure", s.String())

			foundPattern := s.Find(tt.in)

			if tt.shouldFound {
				require.NotEmpty(t, foundPattern)
				require.EqualValues(t, tt.wantTokens, foundPattern.Tokens)
			} else {
				require.Empty(t, foundPattern)
			}
		})
	}
}

func Test_Store_ManySimilarPatterns(t *testing.T) {
	r := NewStore()

	for i := 0; i < 100; i++ {
		if i == 20 {
			r.Add("foo2{p1}foo2{p2}golang")
		}

		if i == 40 {
			r.Add("foo1{p3}foo1{p4}golang")
		}

		r.Add(fmt.Sprintf("%s{p1}%s{p2}golang", RandAZ(4), RandAZ(4)))
	}

	in := "foo1XXXfoo1YYYgolang"

	schema := r.Find(in)
	assert.Len(t, schema.Tokens, 7)
	found, params := schema.Lookup(in)

	assert.True(t, found)
	assert.EqualValues(t, Params{{"p3", "XXX"}, {"p4", "YYY"}}, params)
}

func Benchmark_Store_Lookup_2_2(b *testing.B) {
	r := NewStore()
	r.Add("foo2{p1}foo2{p2}golang")
	r.Add("foo1{p3}foo1{p4}golang")

	b.ReportAllocs()
	b.ResetTimer()
	in := "foo1XXXfoo1YYYgolang"
	for i := 0; i < b.N; i++ {
		r.Find(in)
	}
}

func Benchmark_Store_Lookup_2_102(b *testing.B) {
	r := NewStore()

	for i := 0; i < 100; i++ {
		if i == 20 {
			r.Add("foo2{p1}foo2{p2}golang")
		}

		if i == 40 {
			r.Add("foo1{p3}foo1{p4}golang")
		}

		r.Add(fmt.Sprintf("%s{p1}%s{p2}golang", RandAZ(4), RandAZ(4)))
	}

	b.ReportAllocs()
	b.ResetTimer()
	in := "foo1XXXfoo1YYYgolang"
	for i := 0; i < b.N; i++ {
		r.Find(in)
	}
}

func Test_PatternWithSeparator(t *testing.T) {
	t.Run("ParamBetweenSep", func(t *testing.T) {
		pattern := &Pattern{
			Tokens:    Tokens{StartToken, ConstToken("!"), SeparatorToken("a"), ParameterToken("param"), SeparatorToken("b"), ConstToken("c"), EndToken},
			NumParams: 1,
		}
		t.Log(pattern.String())
		matched, params := pattern.Lookup("!a123bc")
		require.True(t, matched)
		t.Log(params)
	})
}
