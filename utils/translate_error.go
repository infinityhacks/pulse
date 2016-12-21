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
	"regexp"
)

/*
 * Some people, when confronted with a problem, think "I know, I'll use regular
 * expressions". Now they have two problems. -- by Jamie Zawinski
 */

var curlReplacements = []string{
	"^dial tcp: lookup (\\S*) on \\S*: no such host*",
	"DNS lookup failed. $1 could not be resolved (NXDOMAIN).",
}

var curlTranslationTable []*regexp.Regexp

// TranslateError tries to populate field ErrEnglish of a test result
// with a human friendly description of test's error, if any.
//
// Nothing is done if ErrEnglish is already populated.
func TranslateError(result *CombinedResult) {
	// FIXME: do nothing if ErrEnglish is already populated.
	// not yet implemented
}
