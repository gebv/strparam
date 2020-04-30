package strparam

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStore_Lookup(t *testing.T) {
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
	assert.Len(t, schema.Tokens, 5)
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
