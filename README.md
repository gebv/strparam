# strparam

![CI Status](https://github.com/gebv/strparam/workflows/Go/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/gebv/strparam)](https://goreportcard.com/report/github.com/gebv/strparam)
[![codecov](https://codecov.io/gh/gebv/strparam/branch/master/graph/badge.svg)](https://codecov.io/gh/gebv/strparam)

40 times faster аlternative to regex for string matching by pattern and extract params. This is solution as a middle point between simple strings and regular expressions.

Features
* correctly parses UTF-8 characters
* faster than regular expression
* [multiple pattern match](#multiple-pattern-match)

## Introduction

For example. Need to parse the following pattern `foo=(..), baz=(..), golang`. Instead of `..` can be any value.
With regexp, the solution would look something like this.

```golang
in := "foo=(bar), baz=(日本語), golang"
re := regexp.MustCompile(`foo=\((.*)\), baz=\((.*)\), golang`)
re.FindAllStringSubmatch(str, -1)
// [[foo=(bar), baz=(日本語), golang bar 日本語]]
```
[On the playground](https://play.golang.org/p/_ENJU_Mjnty)

Or even like this.

```golang
in := "foo=(bar), baz=(日本語), golang"
re := regexp.MustCompile(`\(([^)]+)\)`)
rex.FindAllStringSubmatch(str, -1)
// [[(bar) bar] [(日本語) 日本語]]
```
[On the playground](https://play.golang.org/p/SSpy7iiINow)

But regular expressions is slow on golang.

Follow the benchmarks for naive solution on regexp (see above) and method `Loockup` for parsed patterns.

```
BenchmarkParamsViaRegexp1
BenchmarkParamsViaRegexp1-4                	   23230	     56140 ns/op	   19258 B/op	       5 allocs/op
BenchmarkParamsViaRegexp2
BenchmarkParamsViaRegexp2-4                	   52396	     23079 ns/op	   28310 B/op	       8 allocs/op
BenchmarkParamsViaStrparam_NumParams2
BenchmarkParamsViaStrparam_NumParams2-4    	  315464	      3467 ns/op	     295 B/op	       1 allocs/op
BenchmarkParamsViaStrparam_NumParams5
BenchmarkParamsViaStrparam_NumParams5-4    	  193682	      5444 ns/op	     296 B/op	       1 allocs/op
BenchmarkParamsViaStrparam_NumParams20
BenchmarkParamsViaStrparam_NumParams20-4   	   72276	     18467 ns/op	     297 B/op	       1 allocs/op
```

Faster solution.

```golang
in := "foo=(bar), baz=(日本語), golang"
s, _ := Parse("foo=({p1}), baz=({p2}), golang")
found, params := s.Lookup(in)
// true [{Name:p1 Value:bar} {Name:p2 Value:日本語}]
```

[On the playground](https://play.golang.org/p/wOS1TUMnl38)

## Multiple pattern match

Performing multiple pattern match for input string. To use a variety of patterns.

```golang
r := NewStore()
r.Add("foo2{p1}foo2{p2}golang")
r.Add("foo1{p3}foo1{p4}golang")

in := "foo1XXXfoo1YYYgolang"

schema, _ := r.Find(in)
found, params := schema.Lookup(in)
```

Follow the benchmarks for method `Store.Find` (without extracting parameters).

```
BenchmarkStore_Lookup_2_2
BenchmarkStore_Lookup_2_2-4                	  255735	      4071 ns/op	     160 B/op	       2 allocs/op
BenchmarkStore_Lookup_2_102
BenchmarkStore_Lookup_2_102-4              	  108709	     12170 ns/op	     160 B/op	       2 allocs/op
```

[On the playground](https://play.golang.org/p/h6u4BHGTsa0)

## Guide

### Installation

```
go get github.com/gebv/strparam
```

### Example

Example for a quick start.

```golang
package main

import (
	"fmt"

	"github.com/gebv/strparam"
)

func main() {
	in := "foo=(bar), baz=(日本語), golang"
	s, _ := strparam.Parse("foo=({p1}), baz=({p2}), golang")
	ok, params := s.Lookup(in)
    fmt.Printf("%v %+v", ok, params)
}

```

[On the playground](https://play.golang.org/p/wOS1TUMnl38)

## How does it work?

Pattern is parse into array of
* tokens with offset information in bytes **for constants**.
* tokens with information of parameter (paremter name and other information).

This pattern `foo=({p1}), baz=({p2}), golang` looks like an array
```
[
    {Mode:begin}
    {Mode:pattern Len:5 Raw:"foo=("} // constant
    {Mode:paremeter Raw:"{p1}"}
    {Mode:pattern Len:8 Raw:"), baz=("}
    {Mode:paremeter Raw:"{p2}"}
    {Mode:pattern Len:9 Raw:"), golang"}
    {Mode:end}
]
```

At the time of parsing the incoming string move around the token array if each token matches. Moving from token to token, we keep the general offset. For parameters, look for the next constant (search window) or end of line.

Prefix-tree is used to store the list of patterns.

For example the follow next patterns:

* `foo{p1}bar`
* `foo{p1}baz`

```
root
    └── foo
        └── {p1}
	        ├── bar
	        └── baz
```

As parsing incoming string we are moving to deep in the tree.

## TODO

- [x] multiple patterns, lookup and extract params
- [ ] extend parameters for internal validators, eg `{paramName required, len=10}`
- [ ] external validators via hooks
- [ ] stream parser

# License

[MIT](LICENSE)
