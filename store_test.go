package strparam

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStore_Lookup(t *testing.T) {
	r := NewStore()
	r.Add("foo2{p1}foo2{p2}golang")
	r.Add("foo1{p3}foo1{p4}golang")

	in := "foo1XXXfoo1YYYgolang"

	schema, err := r.Find(in)
	assert.NoError(t, err)
	found, params := schema.Lookup(in)

	assert.True(t, found)
	assert.EqualValues(t, Params{{"p3", "XXX"}, {"p4", "YYY"}}, params)
}

func BenchmarkStore_Lookup(b *testing.B) {
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
