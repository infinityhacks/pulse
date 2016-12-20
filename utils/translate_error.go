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
	"time"
)

// TranslateError tries to provide a human friendly version
// of a Golang error message.
//
// Parameter testType is one of: TypeDNS, TypeMTR, TypeCurl.
//
// Parameter testDuration is how long the test took to finish.
// It may be used as part of translated message.
//
// Parameter errMsg is the system error returned by Pulse tests.
//
// Returns a human friendly version of the system error,
// or an empty string if a translation is unknown.
func TranslateError(testType int, testDuration time.Duration, errMsg string) string {
	// not yet implemented
	return ""
}
