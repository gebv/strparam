package strparam

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_StoreSingle_ExtendsBasicTests(t *testing.T) {
	tests := []struct {
		pattern    string
		in         string
		found      bool
		wantTokens Tokens
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
			t.Log("[INFO] store", s.String())
			foundSchema := s.Find(tt.in)
			if tt.found {
				require.NotEmpty(t, foundSchema)
				require.EqualValues(t, tt.wantTokens, foundSchema.Tokens)
			} else {
				require.Empty(t, foundSchema)
			}
		})
	}
}

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

			schema := s.Find(tt.in)
			// TODO: check if found
			t.Log("[INFO] found schema:", schema)

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

func Test_StoreMultiple(t *testing.T) {
	s := NewStore()
	err := s.Add("foo2{p1}foo2{p2}golang")
	require.NoError(t, err)
	err = s.Add("foo1{p3}foo1{p4}golang")
	require.NoError(t, err)
	err = s.Add("abc{p5}def{p6}golang")
	require.NoError(t, err)
	t.Log(s.String())
	in := "foo1XXXfoo1YYYgolang"
	schema := s.Find(in)
	require.NotEmpty(t, schema)
	t.Log(schema.Tokens.String())
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
