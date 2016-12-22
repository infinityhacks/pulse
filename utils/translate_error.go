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
	"fmt"
	"regexp"
)

/*
 * Some people, when confronted with a problem, think "I know, I'll use regular
 * expressions". Now they have two problems. -- by Jamie Zawinski
 */

// errorTranslation is a "translation unit".
type errorTranslation struct {
	regexp string
	replacement string
}

// curlErrorTranslations are matched against curl errors during translation.
//
// In replacement strings, the following named parameters
// are also replaced by runtime test values:
// DialTime
var curlErrorTranslations = []errorTranslation{
	errorTranslation{
		".*\\bdial tcp: lookup (\\S+) on \\S*: no such host\\b.*",
		"DNS lookup failed. $1 could not be resolved (NXDOMAIN).",
	},
	errorTranslation{
		".*\\bdial tcp (\\S+): i/o timeout\\b.*",
		"Connection timed out. Agent/client could not connect to $1 within $DialTime seconds.",
	},
}

// curlErrorRegexps contains compiled regexps from curlErrorTranslations.
var curlErrorRegexps []*regexp.Regexp

// TranslateError tries to populate field ErrEnglish of a test result
// with a human friendly description of test's error, if any.
//
// Nothing is done if ErrEnglish is already populated.
func TranslateError(result *CombinedResult) {
	switch result.Type {
		case TypeDNS:
			translateDnsError(result.Result.(*DNSResult))
		case TypeMTR:
			translateMtrError(result.Result.(*MtrResult))
		case TypeCurl:
			translateCurlError(result.Result.(*CurlResult))
	}
}

// translateDnsError tries to populate field ErrEnglish of a DNS test result
// with a human friendly description of test's error, if any.
//
// Nothing is done if ErrEnglish is already populated.
func translateDnsError(result *DNSResult) {
	if result.ErrEnglish != "" {
		return
	}
}

// translateMtrError tries to populate field ErrEnglish of a MTR test result
// with a human friendly description of test's error, if any.
//
// Nothing is done if ErrEnglish is already populated.
func translateMtrError(result *MtrResult) {
	if result.ErrEnglish != "" {
		return
	}
}

// translateCurlError tries to populate field ErrEnglish of a Curl test result
// with a human friendly description of test's error, if any.
//
// Nothing is done if ErrEnglish is already populated.
func translateCurlError(result *CurlResult) {
	if result.ErrEnglish != "" {
		return
	}
	var idx int
	var re *regexp.Regexp
	for idx, re = range curlErrorRegexps {
		if !re.MatchString(result.Err) {
			continue
		}
		timesRe := regexp.MustCompile("\\$({DialTime}|DialTime\\b)")
		repl := timesRe.ReplaceAllString(
			curlErrorTranslations[idx].replacement,
			fmt.Sprintf("%v", result.DialTime.Seconds()),
		)
		result.ErrEnglish = re.ReplaceAllString(result.Err, repl)
		break
	}
}

// Initialize stuff.
func init() {
	curlErrorRegexps = make([]*regexp.Regexp, len(curlErrorTranslations))
	for idx, translation := range curlErrorTranslations {
		curlErrorRegexps[idx] = regexp.MustCompile(translation.regexp)
	}
}
