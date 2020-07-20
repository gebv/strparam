package strparam

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_StoreSingle_Empty(t *testing.T) {
	tests := []struct {
		pattern    string
		in         string
		wantToknes Tokens
	}{
		// {"", "", Tokens{}},
		// {"", "123", Tokens{}},
		// {"123", "", Tokens{}},
		// {"foobar{param}", "", Tokens{}},
		{"{param}foobar", "foobars", Tokens{}},
		// {"foo{param}bar", "", Tokens{}},
		// {"{param}", "", Tokens{StartEndTokens[0], Token{Mode: PARAMETER_PARSED, Param: &Token{Mode: PARAMETER, Raw: "{param}"}}, StartEndTokens[1]}},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%q->%q", tt.pattern, tt.in), func(t *testing.T) {
			s := NewStore()
			s.Add(tt.pattern)
			t.Log("[INFO] store", s.String())
			schema, err := s.Find(tt.in)
			require.NoError(t, err)
			require.EqualValues(t, tt.wantToknes, schema.Tokens)
		})
	}
}

func Test_StoreSingle_FindAndLookup(t *testing.T) {
	// result should be the same as in the case Parse and Loockup for same pattern

	tests := basicCases
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s: %q->%q", tt.name, tt.pattern, tt.in), func(t *testing.T) {
			if tt.wantErr {
				t.Skip("because want error")
			}

			t.Logf("[INFO] pattern = %q, input = %q", tt.pattern, tt.in)

			s := NewStore()
			s.Add(tt.pattern)

			schema, err := s.Find(tt.in)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
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
	s.Add("foo2{p1}foo2{p2}golang")
	s.Add("foo1{p3}foo1{p4}golang")
	s.Add("abc{p5}def{p6}golang")
	t.Log(s.String())
	in := "foo1XXXfoo1YYYgolang"
	schema, err := s.Find(in)
	assert.NoError(t, err)
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

	schema, err := r.Find(in)
	assert.NoError(t, err)
	assert.Len(t, schema.Tokens, 7)
	found, params := schema.Lookup(in)

	assert.True(t, found)
	assert.EqualValues(t, Params{{"p3", "XXX"}, {"p4", "YYY"}}, params)
}

func BenchmarkStore_Lookup_2_2(b *testing.B) {
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

func BenchmarkStore_Lookup_2_102(b *testing.B) {
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
