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

// translateError tries to populate field ErrEnglish of a test result
// with a human friendly description of test's error, if any.
//
// Nothing is done if ErrEnglish is already populated.
func translateError(result *CombinedResult) {
	switch result.Type {
	case TypeDNS:
		results := result.Result.(*DNSResult).Results
		for idx, _ := range results {
			translateDnsError(&results[idx])
		}
	case TypeMTR:
		translateMtrError(result.Result.(*MtrResult))
	case TypeCurl:
		translateCurlError(result.Result.(*CurlResult))
	}
}

// translateDnsError tries to populate ErrEnglish fields of a DNS test result
// with human friendly descriptions of test's errors, if any.
//
// Nothing is done to an already populated ErrEnglish field.
func translateDnsError(result *IndividualDNSResult) {
	if result.ErrEnglish != "" {
		return
	}

	// Some people, when confronted with a problem, think "I know, I'll use
	// regular expressions". Now they have two problems. -- by Jamie Zawinski
	var pattern string
	var re *regexp.Regexp
	var err error

	// Err: "dial udp: lookup some.site.com on 192.168.2.254:53: no such host",
	pattern = ".*\\bdial udp: lookup (\\S+) on \\S*: no such host\\b.*"
	re, err = regexp.Compile(pattern)
	if err == nil && re.MatchString(result.Err) {
		result.ErrEnglish = re.ReplaceAllString(
			result.Err,
			"DNS lookup failed. $1 could not be resolved (NXDOMAIN).",
		)
		return
	}

	// Err: "read udp 192.168.0.13:55155->208.97.182.10:53: i/o timeout",
	pattern = ".*\\bread udp \\S*->(\\S+): i/o timeout\\b.*"
	re, err = regexp.Compile(pattern)
	if err == nil && re.MatchString(result.Err) {
		result.ErrEnglish = re.ReplaceAllString(
			result.Err,
			"DNS lookup timed out. No response from $1 within "+
				inIntegerSeconds(result.Rtt)+
				" seconds.",
		)
		return
	}

	// Err: "read udp 83.169.184.99:53: connection refused",
	// Err: "read udp [2400:cb00:2048:1::c629:d7a2]:53: connection refused",
	pattern = ".*\\bread udp \\[([^]].*)]:([[:digit:]]*): connection refused\\b.*"
	re, err = regexp.Compile(pattern)
	if err == nil && !re.MatchString(result.Err) {
		pattern = ".*\\bread udp ([^:]*):([[:digit:]]*): connection refused\\b.*"
		re, err = regexp.Compile(pattern)
	}
	re, err = regexp.Compile(pattern)
	if err == nil && re.MatchString(result.Err) {
		result.ErrEnglish = re.ReplaceAllString(
			result.Err,
			"DNS lookup refused. $1 refused to accept the DNS query on port ${2}. Maybe nothing is listening on that port or a firewall is blocking.",
		)
		return
	}

	// Err: "dial udp: i/o timeout",
	pattern = ".*\\bdial udp: i/o timeout\\b.*"
	re, err = regexp.Compile(pattern)
	if err == nil && re.MatchString(result.Err) {
		result.ErrEnglish = re.ReplaceAllString(
			result.Err,
			"DNS lookup timed out. Could not resolve "+
				result.Server+
				" to an IP address within "+
				inIntegerSeconds(dnsTimeout)+
				" seconds.",
		)
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

	// Some people, when confronted with a problem, think "I know, I'll use
	// regular expressions". Now they have two problems. -- by Jamie Zawinski
	var pattern string
	var re *regexp.Regexp
	var err error

	// Err: "Get http://lw.cdnplanet.com/static/rum/15kb-image.jpg?t=foo: dial tcp: lookup lw.cdnplanet.com on 8.8.4.4:53: dial udp 8.8.4.4:53: i/o timeout"
	pattern = ".*\\bdial udp (\\S+): i/o timeout\\b.*"
	re, err = regexp.Compile(pattern)
	if err == nil && re.MatchString(result.Err) {
		result.ErrEnglish = re.ReplaceAllString(
			result.Err,
			"DNS lookup timed out. No response from $1 within "+
				inIntegerSeconds(result.DNSTime)+
				" seconds.",
		)
		return
	}

	// Err: "dial tcp: lookup some.site.com on 192.168.1.250:53: server misbehaving"
	pattern = ".*\\bdial tcp: lookup \\S+ on (\\S+): server misbehaving\\b.*"
	re, err = regexp.Compile(pattern)
	if err == nil && re.MatchString(result.Err) {
		result.ErrEnglish = re.ReplaceAllString(
			result.Err,
			"DNS lookup failed. Agent/client can’t reach ${1}.",
		)
		return
	}

	// Err: "Get http://some.site.com/: dial tcp: lookup some.site.com on 192.168.1.1:53: no such host",
	pattern = ".*\\bdial tcp: lookup (\\S+) on \\S*: no such host\\b.*"
	re, err = regexp.Compile(pattern)
	if err == nil && re.MatchString(result.Err) {
		result.ErrEnglish = re.ReplaceAllString(
			result.Err,
			"DNS lookup failed. $1 could not be resolved (NXDOMAIN).",
		)
		return
	}

	// Err: "Get http://8.8.8.8/: dial tcp 8.8.8.8:80: i/o timeout"
	// Err: "Get http://some.site.com/: dial tcp some.site.com:80: i/o timeout",
	// Err: "Get http://2400:cb00:2048:1::c629:d7a2/: dial tcp [2400:cb00:2048:1::c629:d7a2]:80: i/o timeout"
	ipv6 := true
	pattern = ".*\\bdial tcp \\[([^]].*)]:([[:digit:]]*): i/o timeout\\b.*"
	re, err = regexp.Compile(pattern)
	if err == nil && !re.MatchString(result.Err) {
		ipv6 = false
		pattern = ".*\\bdial tcp ([^:]*):([[:digit:]]*): i/o timeout\\b.*"
		re, err = regexp.Compile(pattern)
	}
	if err == nil && re.MatchString(result.Err) {
		var replacement string
		if result.DNSTime == 0 {
			if ipv6 {
				replacement = "Connection timed out. Could not connect to [${1}]:${2} within " +
					inIntegerSeconds(result.DialTime) +
					" seconds."
			} else {
				replacement = "Connection timed out. Could not connect to ${1}:${2} within " +
					inIntegerSeconds(result.DialTime) +
					" seconds."
			}
		} else if result.ConnectTime == 0 {
			replacement = "DNS lookup timed out. Could not resolve $1 within " +
				inIntegerSeconds(result.DialTime) +
				" seconds."
		} else {
			replacement = "Lookup with connection timed out. Could not perform DNS lookup and TCP connection to $1 within " +
				inIntegerSeconds(result.DialTime) +
				" seconds. (DNS lookup " +
				inIntegerMilli(result.DNSTime) +
				"ms, TCP connect " +
				inIntegerMilli(result.ConnectTime) +
				"ms)"
		}
		result.ErrEnglish = re.ReplaceAllString(result.Err, replacement)
		return
	}

	// Err: "Get http://some.site.com/1234/: net/http: timeout awaiting response headers",
	pattern = ".*\\bnet/http: timeout awaiting response headers\\b.*"
	re, err = regexp.Compile(pattern)
	if err == nil && re.MatchString(result.Err) {
		result.ErrEnglish = re.ReplaceAllString(
			result.Err,
			"Request timed out. TCP connection was established but server did not respond to the request within "+
				inIntegerSeconds(responsetimeout)+
				" seconds. (DNS lookup "+
				inIntegerMilli(result.DNSTime)+
				"ms, TCP connect "+
				inIntegerMilli(result.ConnectTime)+
				"ms, TLS handshake "+
				inIntegerMilli(result.TLSTime)+
				"ms)",
		)
		return
	}

	// Err: dial tcp [2400:cb00:2048:1::c629:d7a2]:443: connection refused",
	pattern = ".*\\bdial tcp \\[(\\S+)]:(\\d+): connection refused\\b.*"
	re, err = regexp.Compile(pattern)
	if err == nil && re.MatchString(result.Err) {
		result.ErrEnglish = re.ReplaceAllString(
			result.Err,
			"Connection refused. $1 did not accept the connection on port ${2}.",
		)
		return
	}

	// Err: "dial tcp 203.26.25.4:80: connection refused",
	pattern = ".*\\bdial tcp (\\S+):(\\d+): connection refused\\b.*"
	re, err = regexp.Compile(pattern)
	if err == nil && re.MatchString(result.Err) {
		result.ErrEnglish = re.ReplaceAllString(
			result.Err,
			"Connection refused. $1 did not accept the connection on port ${2}.",
		)
		return
	}

	// Err: "Get https://prod.www-fastly-com.map.fastlylb.net./: x509: certificate is valid for a.ssl.fastly.net, *.a.ssl.fastly.net, rvm.io, not www.nos.nl",
	pattern = ".*\\bx509: certificate is valid for .*, not (\\S+)\\b.*"
	re, err = regexp.Compile(pattern)
	if err == nil && re.MatchString(result.Err) {
		result.ErrEnglish = re.ReplaceAllString(
			result.Err,
			"Certificate is not valid for $1",
		)
		return
	}

	// Err: "Get https://some.site.com/ali-mod/??alicloud-assets-footer/0.0.86/index.css: net/http: TLS handshake timeout",
	pattern = ".*\\bnet/http: TLS handshake timeout\\b.*"
	re, err = regexp.Compile(pattern)
	if err == nil && re.MatchString(result.Err) {
		result.ErrEnglish = re.ReplaceAllString(
			result.Err,
			"TLS handshake timed out.",
		)
		return
	}

}

// inIntegerSeconds formats a Duration to an integer number of seconds.
func inIntegerSeconds(d time.Duration) string {
	return strconv.FormatFloat(d.Seconds(), 'f', 0, 64)
}

// inIntegerMilli formats a Duration to an integer number of milliseconds.
func inIntegerMilli(d time.Duration) string {
	return strconv.FormatFloat(d.Seconds()*1000, 'f', 0, 64)
}
