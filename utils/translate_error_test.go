// Copyright (c) 2016 turbobytes
//
// This file is part of Pulse, a tool to run network diagnostics in a
// distributed manner.
//
// MIT License
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

package pulse

import (
	"testing"
	"time"
)

func TestTranslateError(t *testing.T) {
	type testCase struct {
		testType int
		message string
		expected string
	}
	testCases := []testCase{
		testCase{
			TypeCurl,
			"dial tcp: lookup p.catchpoint.com on 192.168.1.1:53: no such host",
			"DNS lookup failed. p.catchpoint.com could not be resolved (NXDOMAIN).",
		},
	}
	for _, testCase := range testCases {
		received := TranslateError(
			testCase.testType,
			time.Duration(5)*time.Second,
			testCase.message,
		)
		if received != testCase.expected {
			t.Fatalf("translation mismatch: expected \"%s\", got \"%s\"", testCase.expected, received)
		}
	}
}
