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
	"strconv"
	"time"
)

/*
 * Some people, when confronted with a problem, think "I know, I'll use regular
 * expressions". Now they have two problems. -- by Jamie Zawinski
 */

// errorTranslation is a "translation unit".
type errorTranslation struct {
	regexp      string
	replacement string
}

// curlErrorTranslations are matched against curl errors during translation.
//
// In replacement strings, the following named parameters are also recognized:
// DialTimeout (compound timeout for DNS lookup and TCP connection),
// TlsTimeout (timeout for TLS handshake),
// KeepTimeout (timeout for Keep-alive),
// DnsTime (elapsed time during DNS lookup),
// TcpTime (elapsed time during TCP connection),
// DialTime (DnsTime + TcpTime),
// TlsTime (elapsed time during TLS handshake),
// FrbTime (elapsed time awaiting first response byte after sending HTTP request).
//
// In above replacements,
// timeouts are replaced by integer second values such as '5' or '20',
// and elapsed times are replaced by float values with unit identification
// such as '5.2s' or '125ms'.
var curlErrorTranslations = []errorTranslation{
	errorTranslation{
		".*\\bdial tcp: lookup (\\S+) on \\S*: no such host\\b.*",
		"DNS lookup failed. $1 could not be resolved (NXDOMAIN).",
	},
	errorTranslation{
		".*\\bdial tcp (\\S+): i/o timeout\\b.*",
		"Connection timed out. Agent/client could not connect to $1 within $DialTimeout seconds. (DNS lookup ${DnsTime}, TCP connect ${TcpTime})",
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
		replacement := processCurlReplacement(curlErrorTranslations[idx].replacement, result)
		result.ErrEnglish = re.ReplaceAllString(result.Err, replacement)
		break
	}
}

// processCurlReplacement replaces curl named parameters
// (see curlErrorTranslations) in a string.

func processCurlReplacement(repl string, result *CurlResult) string {
	answer := repl
	process := func(re *regexp.Regexp, t time.Duration, isTimeout bool) {
		var val string
		if isTimeout {
			val = strconv.FormatFloat(t.Seconds(), 'f', -1, 64)
		} else {
			val = t.String()
		}
		answer = re.ReplaceAllLiteralString(answer, val)
	}
	process(reDialTimeout, dialtimeout, true)
	process(reTlsTimeout, tlshandshaketimeout, true)
	process(reKeepTimeout, keepalive, true)
	process(reDnsTime, result.DNSTime, false)
	process(reTcpTime, result.ConnectTime, false)
	process(reDialTime, result.DialTime, false)
	process(reTlsTime, result.TLSTime, false)
	process(reFrbTime, result.Ttfb, false)
	return answer
}

// regexps for matching curl named parameters (see curlErrorTranslations).
var (
	reDialTimeout *regexp.Regexp
	reTlsTimeout  *regexp.Regexp
	reKeepTimeout *regexp.Regexp
	reDnsTime     *regexp.Regexp
	reTcpTime     *regexp.Regexp
	reDialTime    *regexp.Regexp
	reTlsTime     *regexp.Regexp
	reFrbTime     *regexp.Regexp
)

// Initialize stuff.
func init() {
	// Compile error regexps for Curl tests.
	curlErrorRegexps = make([]*regexp.Regexp, len(curlErrorTranslations))
	for idx, translation := range curlErrorTranslations {
		curlErrorRegexps[idx] = regexp.MustCompile(translation.regexp)
	}
	// paramRegexp returns a regexp that matches a named parameter
	// in forms $... or ${...}
	paramRegexp := func(paramName string) *regexp.Regexp {
		return regexp.MustCompile("\\$({" + paramName + "}|" + paramName + "\\b)")
	}
	// Compile regexps for Curl named parameters.
	reDialTimeout = paramRegexp("DialTimeout")
	reTlsTimeout = paramRegexp("TlsTimeout")
	reKeepTimeout = paramRegexp("KeepTimeout")
	reDnsTime = paramRegexp("DnsTime")
	reTcpTime = paramRegexp("TcpTime")
	reDialTime = paramRegexp("DialTime")
	reTlsTime = paramRegexp("TlsTime")
	reFrbTime = paramRegexp("FrbTime")
}
