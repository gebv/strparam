# strparam

![CI Status](https://github.com/gebv/strparam/workflows/Go/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/gebv/strparam)](https://goreportcard.com/report/github.com/gebv/strparam)
[![codecov](https://codecov.io/gh/gebv/strparam/branch/master/graph/badge.svg)](https://codecov.io/gh/gebv/strparam)

40 times faster аlternative to regex for string matching by pattern and extract params.

Features
* correctrly parse UTF-8 characters
* faster than regular expression
* [lookup on multiple patterns](#lookup-on-multiple-patterns)

## Introduction

Pattern example `foo=(..), baz=(..), golang`. Instead of `..` can be any value.

With regexp, the solution would look something like this.

```golang
in := "foo=(bar), baz=(日本語), golang"
re := regexp.MustCompile(`foo=\((.*)\), baz=\((.*)\), golang`)
re.FindAllStringSubmatch(str, -1)
// [[foo=(bar), baz=(日本語), golang bar 日本語]]
```
[On playground](https://play.golang.org/p/_ENJU_Mjnty)

Or like this.

```golang
in := "foo=(bar), baz=(日本語), golang"
re := regexp.MustCompile(`\(([^)]+)\)`)
rex.FindAllStringSubmatch(str, -1)
// [[(bar) bar] [(日本語) 日本語]]
```
[On playground](https://play.golang.org/p/SSpy7iiINow)

But regular expressions is slow on golang.

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

[On playground](https://play.golang.org/p/wOS1TUMnl38)

## Lookup on multiple patterns

```golang
r := NewStore()
r.Add("foo2{p1}foo2{p2}golang")
r.Add("foo1{p3}foo1{p4}golang")

in := "foo1XXXfoo1YYYgolang"

schema, _ := r.Find(in)
found, params := schema.Lookup(in)
```

[On playground](https://play.golang.org/p/h6u4BHGTsa0)

## Guide

### Installation

```
go get github.com/gebv/strparam
```

### Example

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

[On playground](https://play.golang.org/p/wOS1TUMnl38)

## How does it work?

// TODO:

## TODO

- [x] multiple patterns, lookup and extract params
- [ ] extend parameters for internal validators, eg `{paramName required, len=10}`
- [ ] external validators via hooks
- [ ] stream parser

# License

[MIT](LICENSE)
