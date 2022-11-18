// Copied from https://github.com/junk1tm/assert/blob/v0.1.0/dotimport/alias.go
//
// Copyright (c) 2022 junk1tm
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

// Package dotimport provides type aliases for the parent [assert] package. It
// is intended to be imported using dot syntax so that [E] and [F] can be used
// as if they were local types.
//
//	package foo_test
//
//	import (
//		"testing"
//
//		"assert"
//		. "assert/dotimport"
//	)
//
//	func TestFoo(t *testing.T) {
//		assert.NoErr[E](t, foo.Foo())
//	}
package dotimport

import "github.com/go-simpler/env/assert"

type (
	E = assert.E
	F = assert.F
)
