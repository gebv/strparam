package strparam

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var basicCases = []struct {
	name         string
	pattern      string
	in           string
	want         Params
	found        bool
	wantErr      bool
	schemaTokens Tokens
}{
	{"empty", "", "", Params{}, true, false, EmptySchema},
	{"empty", "", "qwe", nil, false, false, EmptySchema},

	{"p", "qwe", "", nil, false, false, Tokens{StartToken, PatternToken("qwe"), EndToken}},
	{"p", "qwe", "qwe", Params{}, true, false, Tokens{StartToken, PatternToken("qwe"), EndToken}},
	{"p", "qwe", "qwe123", nil, false, false, Tokens{StartToken, PatternToken("qwe"), EndToken}},
	{"p", "qwe", "123qwe", nil, false, false, Tokens{StartToken, PatternToken("qwe"), EndToken}},
	{"p", "qwe", "qw123e", nil, false, false, Tokens{StartToken, PatternToken("qwe"), EndToken}},

	{"{}", "{qwe}", "123", Params{{"qwe", "123"}}, true, false, Tokens{StartToken, ParameterToken("qwe"), EndToken}},
	{"{}", "{qwe}", "", Params{{"qwe", ""}}, true, false, Tokens{StartToken, ParameterToken("qwe"), EndToken}},

	{"{}p", "{qwe}foo", "", nil, false, false, Tokens{StartToken, ParameterToken("qwe"), PatternToken("foo"), EndToken}},
	{"{}p", "{qwe}foo", "123", nil, false, false, Tokens{StartToken, ParameterToken("qwe"), PatternToken("foo"), EndToken}},
	{"{}p", "{qwe}foo", "123foo", Params{{"qwe", "123"}}, true, false, Tokens{StartToken, ParameterToken("qwe"), PatternToken("foo"), EndToken}},
	{"{}p", "{qwe}foo", "123foo123", nil, false, false, Tokens{StartToken, ParameterToken("qwe"), PatternToken("foo"), EndToken}},
	{"{}p", "{qwe}foo", "foo123", nil, false, false, Tokens{StartToken, ParameterToken("qwe"), PatternToken("foo"), EndToken}},
	{"{}p", "{qwe}foo", "foo", Params{{"qwe", ""}}, true, false, Tokens{StartToken, ParameterToken("qwe"), PatternToken("foo"), EndToken}},

	{"p{}", "foo{qwe}", "", nil, false, false, Tokens{StartToken, PatternToken("foo"), ParameterToken("qwe"), EndToken}},
	{"p{}", "foo{qwe}", "123", nil, false, false, Tokens{StartToken, PatternToken("foo"), ParameterToken("qwe"), EndToken}},
	{"p{}", "foo{qwe}", "foo123", Params{{"qwe", "123"}}, true, false, Tokens{StartToken, PatternToken("foo"), ParameterToken("qwe"), EndToken}},
	{"p{}", "foo{qwe}", "123foo123", nil, false, false, Tokens{StartToken, PatternToken("foo"), ParameterToken("qwe"), EndToken}},
	{"p{}", "foo{qwe}", "123foo", nil, false, false, Tokens{StartToken, PatternToken("foo"), ParameterToken("qwe"), EndToken}},
	{"p{}", "foo{qwe}", "foo", Params{{"qwe", ""}}, true, false, Tokens{StartToken, PatternToken("foo"), ParameterToken("qwe"), EndToken}},

	{"p{}p", "foo{qwe}bar", "", nil, false, false, Tokens{StartToken, PatternToken("foo"), ParameterToken("qwe"), PatternToken("bar"), EndToken}},
	{"p{}p", "foo{qwe}bar", "foo", nil, false, false, Tokens{StartToken, PatternToken("foo"), ParameterToken("qwe"), PatternToken("bar"), EndToken}},
	{"p{}p", "foo{qwe}bar", "bar", nil, false, false, Tokens{StartToken, PatternToken("foo"), ParameterToken("qwe"), PatternToken("bar"), EndToken}},
	{"p{}p", "foo{qwe}bar", "barfoo", nil, false, false, Tokens{StartToken, PatternToken("foo"), ParameterToken("qwe"), PatternToken("bar"), EndToken}},
	{"p{}p", "foo{qwe}bar", "barfoo123", nil, false, false, Tokens{StartToken, PatternToken("foo"), ParameterToken("qwe"), PatternToken("bar"), EndToken}},
	{"p{}p", "foo{qwe}bar", "foo123", nil, false, false, Tokens{StartToken, PatternToken("foo"), ParameterToken("qwe"), PatternToken("bar"), EndToken}},
	{"p{}p", "foo{qwe}bar", "123bar", nil, false, false, Tokens{StartToken, PatternToken("foo"), ParameterToken("qwe"), PatternToken("bar"), EndToken}},
	{"p{}p", "foo{qwe}bar", "foobar", Params{{"qwe", ""}}, true, false, Tokens{StartToken, PatternToken("foo"), ParameterToken("qwe"), PatternToken("bar"), EndToken}},
	{"p{}p", "foo{qwe}bar", "foo123bar", Params{{"qwe", "123"}}, true, false, Tokens{StartToken, PatternToken("foo"), ParameterToken("qwe"), PatternToken("bar"), EndToken}},
	{"p{}p", "foo{qwe}bar", "123foo123bar", nil, false, false, Tokens{StartToken, PatternToken("foo"), ParameterToken("qwe"), PatternToken("bar"), EndToken}},
	{"p{}p", "foo{qwe}bar", "foo123bar123", nil, false, false, Tokens{StartToken, PatternToken("foo"), ParameterToken("qwe"), PatternToken("bar"), EndToken}},
	{"p{}p", "foo{qwe}bar", "foobar123", nil, false, false, Tokens{StartToken, PatternToken("foo"), ParameterToken("qwe"), PatternToken("bar"), EndToken}},
	{"p{}p", "foo{qwe}bar", "123foobar", nil, false, false, Tokens{StartToken, PatternToken("foo"), ParameterToken("qwe"), PatternToken("bar"), EndToken}},

	{"utf8pattern", "foo{p1}日本語{p2}baz", "fooAAA日本語BBBbaz", Params{{"p1", "AAA"}, {"p2", "BBB"}}, true, false, Tokens{StartToken, PatternToken("foo"), ParameterToken("p1"), PatternToken("日本語"), ParameterToken("p2"), PatternToken("baz"), EndToken}},
	{"utf8param", "foo{p1}bar{p2}baz", "foo日本語barСЫРbaz", Params{{"p1", "日本語"}, {"p2", "СЫР"}}, true, false, Tokens{StartToken, PatternToken("foo"), ParameterToken("p1"), PatternToken("bar"), ParameterToken("p2"), PatternToken("baz"), EndToken}},

	{"invalidParse", "{foo}{bar}", "", nil, false, true, nil},
	{"invalidParse", "{foo}{bar}", "", nil, false, true, nil},
	{"invalidParse", "{}{bar}", "", nil, false, true, nil},
	{"invalidParse", "{foo}{bar", "", nil, false, true, nil},
	{"invalidParse", "{foo}{", "", nil, false, true, nil},
	{"invalidParse", "{foo}{{", "", nil, false, true, nil},
	{"invalidParse", "{foo}{}", "", nil, false, true, nil},
	{"invalidParse", "{}{}", "", nil, false, true, nil},
	{"invalidParse", "{", "", nil, false, true, nil},
	{"invalidParse", "{{}", "", nil, false, true, nil},

	{"{}p{}", "{p1}qw{p2}", "", nil, false, false, Tokens{StartToken, ParameterToken("p1"), PatternToken("qw"), ParameterToken("p2"), EndToken}},
	{"{}p{}", "{p1}qw{p2}", "123", nil, false, false, Tokens{StartToken, ParameterToken("p1"), PatternToken("qw"), ParameterToken("p2"), EndToken}},
	{"{}p{}", "{p1}qw{p2}", "q", nil, false, false, Tokens{StartToken, ParameterToken("p1"), PatternToken("qw"), ParameterToken("p2"), EndToken}},
	{"{}p{}", "{p1}qw{p2}", "w", nil, false, false, Tokens{StartToken, ParameterToken("p1"), PatternToken("qw"), ParameterToken("p2"), EndToken}},
	{"{}p{}", "{p1}qw{p2}", "qw", Params{{"p1", ""}, {"p2", ""}}, true, false, Tokens{StartToken, ParameterToken("p1"), PatternToken("qw"), ParameterToken("p2"), EndToken}},
	{"{}p{}", "{p1}qw{p2}", "qw123", Params{{"p1", ""}, {"p2", "123"}}, true, false, Tokens{StartToken, ParameterToken("p1"), PatternToken("qw"), ParameterToken("p2"), EndToken}},
	{"{}p{}", "{p1}qw{p2}", "w123", nil, false, false, Tokens{StartToken, ParameterToken("p1"), PatternToken("qw"), ParameterToken("p2"), EndToken}},
	{"{}p{}", "{p1}qw{p2}", "qw123456", Params{{"p1", ""}, {"p2", "123456"}}, true, false, Tokens{StartToken, ParameterToken("p1"), PatternToken("qw"), ParameterToken("p2"), EndToken}},
	{"{}p{}", "{p1}qw{p2}", "123qw", Params{{"p1", "123"}, {"p2", ""}}, true, false, Tokens{StartToken, ParameterToken("p1"), PatternToken("qw"), ParameterToken("p2"), EndToken}},
	{"{}p{}", "{p1}qw{p2}", "123q", nil, false, false, Tokens{StartToken, ParameterToken("p1"), PatternToken("qw"), ParameterToken("p2"), EndToken}},
	{"{}p{}", "{p1}qw{p2}", "456123qw", Params{{"p1", "456123"}, {"p2", ""}}, true, false, Tokens{StartToken, ParameterToken("p1"), PatternToken("qw"), ParameterToken("p2"), EndToken}},

	{"p{}p{}p", "foo{p1}bar{p2}baz", "", nil, false, false, Tokens{StartToken, PatternToken("foo"), ParameterToken("p1"), PatternToken("bar"), ParameterToken("p2"), PatternToken("baz"), EndToken}},
	{"p{}p{}p", "foo{p1}bar{p2}baz", "foo", nil, false, false, Tokens{StartToken, PatternToken("foo"), ParameterToken("p1"), PatternToken("bar"), ParameterToken("p2"), PatternToken("baz"), EndToken}},
	{"p{}p{}p", "foo{p1}bar{p2}baz", "foobaz", nil, false, false, Tokens{StartToken, PatternToken("foo"), ParameterToken("p1"), PatternToken("bar"), ParameterToken("p2"), PatternToken("baz"), EndToken}},
	{"p{}p{}p", "foo{p1}bar{p2}baz", "bar", nil, false, false, Tokens{StartToken, PatternToken("foo"), ParameterToken("p1"), PatternToken("bar"), ParameterToken("p2"), PatternToken("baz"), EndToken}},
	{"p{}p{}p", "foo{p1}bar{p2}baz", "barbaz", nil, false, false, Tokens{StartToken, PatternToken("foo"), ParameterToken("p1"), PatternToken("bar"), ParameterToken("p2"), PatternToken("baz"), EndToken}},
	{"p{}p{}p", "foo{p1}bar{p2}baz", "foobar", nil, false, false, Tokens{StartToken, PatternToken("foo"), ParameterToken("p1"), PatternToken("bar"), ParameterToken("p2"), PatternToken("baz"), EndToken}},
	{"p{}p{}p", "foo{p1}bar{p2}baz", "foobarbaz", Params{{"p1", ""}, {"p2", ""}}, true, false, Tokens{StartToken, PatternToken("foo"), ParameterToken("p1"), PatternToken("bar"), ParameterToken("p2"), PatternToken("baz"), EndToken}},
	{"p{}p{}p", "foo{p1}bar{p2}baz", "foo123barbaz", Params{{"p1", "123"}, {"p2", ""}}, true, false, Tokens{StartToken, PatternToken("foo"), ParameterToken("p1"), PatternToken("bar"), ParameterToken("p2"), PatternToken("baz"), EndToken}},
	{"p{}p{}p", "foo{p1}bar{p2}baz", "foo123bar456baz", Params{{"p1", "123"}, {"p2", "456"}}, true, false, Tokens{StartToken, PatternToken("foo"), ParameterToken("p1"), PatternToken("bar"), ParameterToken("p2"), PatternToken("baz"), EndToken}},
	{"p{}p{}p", "foo{p1}bar{p2}baz", "foo123bar456baz789", nil, false, false, Tokens{StartToken, PatternToken("foo"), ParameterToken("p1"), PatternToken("bar"), ParameterToken("p2"), PatternToken("baz"), EndToken}},
	{"p{}p{}p", "foo{p1}bar{p2}baz", "foobar456baz789", nil, false, false, Tokens{StartToken, PatternToken("foo"), ParameterToken("p1"), PatternToken("bar"), ParameterToken("p2"), PatternToken("baz"), EndToken}},
	{"p{}p{}p", "foo{p1}bar{p2}baz", "foobarbaz789", nil, false, false, Tokens{StartToken, PatternToken("foo"), ParameterToken("p1"), PatternToken("bar"), ParameterToken("p2"), PatternToken("baz"), EndToken}},
	{"p{}p{}p", "foo{p1}bar{p2}baz", "456foo123bar456baz", nil, false, false, Tokens{StartToken, PatternToken("foo"), ParameterToken("p1"), PatternToken("bar"), ParameterToken("p2"), PatternToken("baz"), EndToken}},
	{"p{}p{}p", "foo{p1}bar{p2}baz", "456foo123barbaz", nil, false, false, Tokens{StartToken, PatternToken("foo"), ParameterToken("p1"), PatternToken("bar"), ParameterToken("p2"), PatternToken("baz"), EndToken}},
	{"p{}p{}p", "foo{p1}bar{p2}baz", "456foobarbaz", nil, false, false, Tokens{StartToken, PatternToken("foo"), ParameterToken("p1"), PatternToken("bar"), ParameterToken("p2"), PatternToken("baz"), EndToken}},
	{"p{}p{}p", "foo{p1}bar{p2}baz", "456foobarbaz123", nil, false, false, Tokens{StartToken, PatternToken("foo"), ParameterToken("p1"), PatternToken("bar"), ParameterToken("p2"), PatternToken("baz"), EndToken}},
	{"p{}p{}p", "foo{p1}bar{p2}baz", "456foo123barbaz123", nil, false, false, Tokens{StartToken, PatternToken("foo"), ParameterToken("p1"), PatternToken("bar"), ParameterToken("p2"), PatternToken("baz"), EndToken}},
	{"p{}p{}p", "foo{p1}bar{p2}baz", "456foo123bar123baz123", nil, false, false, Tokens{StartToken, PatternToken("foo"), ParameterToken("p1"), PatternToken("bar"), ParameterToken("p2"), PatternToken("baz"), EndToken}},

	// // https://github.com/gebv/strparam/issues/3
	{"issues#3", "{{bar}", "{123", Params{{"bar", "123"}}, true, false, Tokens{StartToken, PatternToken("{"), ParameterToken("bar"), EndToken}},
}

func Test_Parse(t *testing.T) {
	tests := []struct {
		name         string
		pattern      string
		in           string
		want         Params
		found        bool
		wantErr      bool
		schemaTokens Tokens
	}(basicCases)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("[INFO] pattern = %q, input = %q", tt.pattern, tt.in)

			schema, err := Parse(tt.pattern)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				require.Nil(t, schema)
			} else {
				require.EqualValues(t, tt.schemaTokens, schema.Tokens)
			}

			if err == nil {
				t.Log("[INFO] found schema:", schema)
				for _, token := range schema.Tokens {
					if token.Mode == PATTERN {
						assert.True(t, token.Len > 0, "pattern must not be empty")
						assert.True(t, token.Raw != "pattern must not be empty")
					}
				}
			}

			found, params := schema.Lookup(tt.in)
			if found != tt.found {
				t.Errorf("Loockup().found = %v, want %v", found, tt.found)
			}
			require.EqualValues(t, tt.want, params)
		})
	}
}
func Test_ParseAndLookup_EmptySchema(t *testing.T) {
	s := &PatternSchema{Tokens: []Token{}, NumParams: 0}
	t.Run("emptyListTokensNotMatchedInEmpty", func(t *testing.T) {
		found, params := s.Lookup("")
		assert.False(t, found)
		assert.Empty(t, params)
	})
	t.Run("emptyListTokensNotMatchedInAnything", func(t *testing.T) {
		found, params := s.Lookup("123")
		assert.False(t, found)
		assert.Empty(t, params)
	})

	s, err := Parse("")
	require.NoError(t, err)
	require.Len(t, s.Tokens, 2)
	require.Equal(t, EmptySchema, s.Tokens)

	t.Run("emptySchemaMatchedInEmpty", func(t *testing.T) {
		found, params := s.Lookup("")
		assert.True(t, found)
		assert.Empty(t, params)
	})
	t.Run("emptySchemaNotMatchedInAnything", func(t *testing.T) {
		found, params := s.Lookup("123")
		assert.False(t, found)
		assert.Empty(t, params)
	})

}

func TestDemoRegexp(t *testing.T) {
	in := "foo=(bar), baz=(日本語), golang"
	t.Run("regexp1", func(t *testing.T) {
		re := regexp.MustCompile(`foo=\((.*)\), baz=\((.*)\), golang`)
		res := re.FindAllStringSubmatch(in, -1)
		t.Logf("%+v", res)
		assert.EqualValues(t, [][]string{{"foo=(bar), baz=(日本語), golang", "bar", "日本語"}}, res)
	})

	t.Run("regexp2", func(t *testing.T) {
		re := regexp.MustCompile(`\(([^)]+)\)`)
		res := re.FindAllStringSubmatch(in, -1)
		t.Logf("%+v", res)
		assert.EqualValues(t, [][]string{{"(bar)", "bar"}, {"(日本語)", "日本語"}}, res)
	})

	t.Run("strparam", func(t *testing.T) {
		in := "foo=(bar), baz=(日本語), golang"
		s, err := Parse("foo=({p1}), baz=({p2}), golang")
		assert.NoError(t, err)
		ok, params := s.Lookup(in)
		t.Logf("%v %+v", ok, params)
		assert.True(t, ok)
		assert.EqualValues(t, Params{{"p1", "bar"}, {"p2", "日本語"}}, params)
	})

}

func BenchmarkParamsViaRegexp1(b *testing.B) {
	str := "foo=(bar), baz=(日本語), golang"
	rex := regexp.MustCompile(`foo=\((.*)\), baz=\((.*)\), golang`)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rex.FindAllStringSubmatch(str, -1)
	}
}

func BenchmarkParamsViaRegexp2(b *testing.B) {
	str := "foo=(bar), baz=(日本語), golang"
	rex := regexp.MustCompile(`\(([^)]+)\)`)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rex.FindAllStringSubmatch(str, -1)
	}
}

func BenchmarkParamsViaStrparam_NumParams2(b *testing.B) {
	in := "foo=(bar), baz=(日本語), golang"
	s, _ := Parse("foo=({p1}), baz=({p2}), golang")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Lookup(in)
	}
}

func BenchmarkParamsViaStrparam_NumParams5(b *testing.B) {
	in := "foo1=(bar), baz2=(日本語), foo3=(bar), baz4=(日本語), foo5=(bar) golang"
	s, _ := Parse("foo1=({p1}), baz2=({p2}), foo3=({p3}), baz4=({p4}), foo5=({p5}) golang")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Lookup(in)
	}
}

func BenchmarkParamsViaStrparam_NumParams20(b *testing.B) {
	in := "foo1=(bar), baz2=(日本語), foo3=(bar), baz4=(日本語), foo5=(bar), baz6=(日本語), foo7=(bar), baz8=(日本語), foo9=(bar), baz10=(日本語), foo11=(bar), baz12=(日本語), foo13=(bar), baz14=(日本語), foo15=(bar), baz16=(日本語), foo17=(bar), baz18=(日本語), foo19=(bar), baz20=(日本語) golang"
	s, _ := Parse("foo1=({p1}), baz2=({p2}), foo3=({p3}), baz4=({p4}), foo5=({p5}), baz6=({p6}), foo7=({p7}), baz8=({p8}), foo9=({p9}), baz10=({p10}), foo11=({p11}), baz12=({p12}), foo13=({p13}), baz14=({p14}), foo15=({p15}), baz16=({p16}), foo17=({p17}), baz18=({p18}), foo19=({p19}), baz20=({p20}) golang")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Lookup(in)
	}
}

func TestPatternSchema_Name(t *testing.T) {
	tests := []struct {
		name   string
		Tokens Tokens
		want   string
	}{
		{name: "empty"},
		{"invalid", Tokens{{Mode: PATTERN, Raw: "123"}}, ""},
		{"invalid", Tokens{{Mode: BEGINLINE, Raw: "123"}}, ""},
		{"invalid", Tokens{{Mode: ENDLINE, Raw: "123"}}, ""},
		{"", Tokens{{Mode: BEGINLINE}, {Mode: ENDLINE}}, ""},
		{"", Tokens{{Mode: BEGINLINE}, {Mode: ENDLINE, Raw: "foobar"}}, "foobar"},
		{"", Tokens{{Mode: BEGINLINE}, {Mode: PATTERN, Raw: "123"}, {Mode: ENDLINE, Raw: "foobar"}}, "foobar"},
		{"invalid", Tokens{{Mode: BEGINLINE}, {Mode: PATTERN, Raw: "123"}}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := PatternSchema{
				Tokens: tt.Tokens,
			}
			if got := s.Name(); got != tt.want {
				t.Errorf("PatternSchema.Name() = %v, want %v", got, tt.want)
			}
		})
	}
}
