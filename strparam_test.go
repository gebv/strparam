package strparam

import (
	"reflect"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		in      string
		want    Params
		found   bool
		wantErr bool
	}{
		{"empty", "", "", Params{}, true, false},
		{"withoutParams", "foobar", "foo123bar", nil, false, false},
		{"simple1", "foo{p1}bar", "foo123bar", Params{{"p1", "123"}}, true, false},
		{"utf8pattern", "foo{p1}日本語{p2}baz", "fooAAA日本語BBBbaz", Params{{"p1", "AAA"}, {"p2", "BBB"}}, true, false},
		{"utf8param", "foo{p1}bar{p2}baz", "foo日本語barСЫРbaz", Params{{"p1", "日本語"}, {"p2", "СЫР"}}, true, false},
		{"issues#1", "#snippet-{boundary}", "foobar", nil, false, false},
		{"issues#1", "verylongpattern-{p1}", "smallinput", nil, false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema, err := Parse(tt.pattern)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
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
