package strparam

import (
	"reflect"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var basicCases = []struct {
	name    string
	pattern string
	in      string
	want    Params
	found   bool
	wantErr bool
}{
	{"empty", "", "", Params{}, true, false},
	{"empty", "", "qwe", nil, false, false},
	{"empty", "qwe", "", nil, false, false},
	{"empty", "qwe", "qwe", Params{}, true, false},
	{"empty", "qwe", "qwe123", nil, false, false},
	{"empty", "qwe", "123qwe", nil, false, false},
	{"empty", "qwe", "qw123e", nil, false, false},
	{"allAsParam", "{qwe}", "123", Params{{"qwe", "123"}}, true, false},
	{"allAsParam", "{qwe}", "", Params{{"qwe", ""}}, true, false},
	{"onlyTwoParams", "{foo}{bar}", "", nil, false, true},
	{"onlyTwoParams", "{foo}{bar}", "123", nil, false, true},
	{"onlyTwoParams", "foo{foo}{bar}", "foo123", nil, false, true},
	{"onlyTwoParams", "{foo}foo{bar}", "123foo", Params{{"foo", "123"}, {"bar", ""}}, true, false},
	{"onlyTwoParams", "{foo}foo{bar}", "foo456", Params{{"foo", ""}, {"bar", "456"}}, true, false},
	{"withoutParams", "foobar", "foo123bar", nil, false, false},
	{"simple1", "foo{p1}bar", "foo123bar", Params{{"p1", "123"}}, true, false},
	{"simple1-empytparamvalue", "foo{p1}bar", "foobar", Params{{"p1", ""}}, true, false},
	{"utf8pattern", "foo{p1}日本語{p2}baz", "fooAAA日本語BBBbaz", Params{{"p1", "AAA"}, {"p2", "BBB"}}, true, false},
	{"utf8param", "foo{p1}bar{p2}baz", "foo日本語barСЫРbaz", Params{{"p1", "日本語"}, {"p2", "СЫР"}}, true, false},
	{"issues#1", "#snippet-{boundary}", "foobar", nil, false, false},
	{"issues#1", "verylongpattern-{p1}", "smallinput", nil, false, false},
	{"issues#2", "{v1}fooobar{v2}", "1fooobar2", Params{{"v1", "1"}, {"v2", "2"}}, true, false},
	{"issues#2", "{v1}fooobar{v2}", "1fooobar", Params{{"v1", "1"}, {"v2", ""}}, true, false},
	{"issues#2", "{v1}fooobar", "111fooobar", Params{{"v1", "111"}}, true, false},
	{"issues#2", "{v1}fooobar", "fooobar", Params{{"v1", ""}}, true, false},
	{"issues#2", "fooobar{v2}", "fooobar222", Params{{"v2", "222"}}, true, false},
	{"issues#2", "fooobar{v2}", "fooobar", Params{{"v2", ""}}, true, false},
}

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		in      string
		want    Params
		found   bool
		wantErr bool
	}(basicCases)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("[INFO] pattern = %q, input = %q", tt.pattern, tt.in)

			schema, err := Parse(tt.pattern)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				t.Logf("[INFO] found schema %q", schema.Tokens)
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
			if !reflect.DeepEqual(params, tt.want) {
				t.Errorf("Loockup().params = %v, want %v", params, tt.want)
			}
		})
	}
}
func TestParseAndLookup_EmptySchema(t *testing.T) {
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
	require.Equal(t, s.Tokens[0].Mode, BEGINLINE)
	require.Equal(t, s.Tokens[1].Mode, ENDLINE)

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
