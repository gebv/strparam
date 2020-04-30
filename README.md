# strparam

![CI Status](https://github.com/gebv/strparam/workflows/Go/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/gebv/strparam)](https://goreportcard.com/report/github.com/gebv/strparam)
[![codecov](https://codecov.io/gh/gebv/strparam/branch/master/graph/badge.svg)](https://codecov.io/gh/gebv/strparam)

40 times faster аlternative to regex for string matching by pattern and extract params.

Features
* correctrly parse UTF-8 characters
* faster regexp

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
BenchmarkParamsViaRegexp1-4    	   22818	     50851 ns/op	   18994 B/op	       5 allocs/op
BenchmarkParamsViaRegexp2
BenchmarkParamsViaRegexp2-4    	   47041	     24834 ns/op	   28218 B/op	       8 allocs/op
BenchmarkParamsViaStrparam
BenchmarkParamsViaStrparam-4   	  474166	      2248 ns/op	      64 B/op	       1 allocs/op
```

Faster solution.

```golang
in := "foo=(bar), baz=(日本語), golang"
s, _ := Parse("foo=({p1}), baz=({p2}), golang")
found, params := s.Lookup(in)
// true [{Name:p1 Value:bar} {Name:p2 Value:日本語}]
```

[On playground](https://play.golang.org/p/wOS1TUMnl38)

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

- [ ] multiple patterns, lookup and extract params
- [ ] extend parameters for internal validators, eg `{paramName required, len=10}`
- [ ] external validators via hooks
- [ ] stream parser

# License

[MIT](LICENSE)
