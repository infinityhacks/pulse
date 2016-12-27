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
		testResult CombinedResult
		expected   string
	}
	ms50 := time.Millisecond * 50
	ms128 := time.Millisecond * 128
	ms429 := time.Millisecond * 429
	ms2041 := time.Millisecond * 2041
	ms10050 := time.Millisecond * 10050
	testCases := []testCase{
		testCase{
			CombinedResult{
				Type: TypeCurl,
				Result: &CurlResult{
					Err: "dial tcp: lookup p.catchpoint.com on 192.168.1.1:53: no such host",
				},
			},
			"DNS lookup failed. p.catchpoint.com could not be resolved (NXDOMAIN).",
		},
		testCase{
			CombinedResult{
				Type: TypeCurl,
				Result: &CurlResult{
					Err:         "Get http://8.8.8.8/: dial tcp 8.8.8.8:80: i/o timeout",
					DNSTime:     ms128,
					ConnectTime: dialtimeout - ms128,
					DialTime:    dialtimeout,
				},
			},
			"Connection timed out. Agent/client could not connect to 8.8.8.8:80 within 15 seconds. (DNS lookup 128ms, TCP connect 14.872s)",
		},
		testCase{
			CombinedResult{
				Type: TypeCurl,
				Result: &CurlResult{
					Err:         "net/http: timeout awaiting response headers",
					DNSTime:     ms128,
					ConnectTime: ms2041,
					DialTime:    ms128 + ms2041,
					TLSTime:     ms429,
					Ttfb:        ms10050,
				},
			},
			"Request timed out. TCP connection was established but server did not respond to the request within 10 seconds. (DNS lookup 128ms, TCP connect 2.041s, TLS handshake 429ms)",
		},
		testCase{
			CombinedResult{
				Type: TypeCurl,
				Result: &CurlResult{
					Err:      "Get http://lw.cdnplanet.com/static/rum/15kb-image.jpg?t=foo: dial tcp: lookup lw.cdnplanet.com on 8.8.4.4:53: dial udp 8.8.4.4:53: i/o timeout",
					DNSTime:  dialtimeout + ms50,
					DialTime: dialtimeout + ms50,
				},
			},
			"DNS lookup timed out. No response from 8.8.4.4:53 within 15 seconds.",
		},
		testCase{
			CombinedResult{
				Type: TypeCurl,
				Result: &CurlResult{
					Err: "dial tcp 203.26.25.4:80: connection refused",
				},
			},
			"Connection refused. 203.26.25.4 did not accept the connection on port 80.",
		},
		testCase{
			CombinedResult{
				Type: TypeCurl,
				Result: &CurlResult{
					Err: "dial tcp [2400:cb00:2048:1::c629:d7a2]:443: connection refused",
				},
			},
			"Connection refused. 2400:cb00:2048:1::c629:d7a2 did not accept the connection on port 443.",
		},
		testCase{
			CombinedResult{
				Type: TypeCurl,
				Result: &CurlResult{
					Err: "dial tcp: lookup cdn.albel.li on 192.168.1.250:53: server misbehaving",
				},
			},
			"DNS lookup failed. Agent/client can’t reach 192.168.1.250:53.",
		},
	}
	for _, testCase := range testCases {
		TranslateError(&testCase.testResult)
		translated := testCase.testResult.Result.(*CurlResult).ErrEnglish
		if translated != testCase.expected {
			t.Fatalf("translation mismatch: expected \"%s\", got \"%s\"", testCase.expected, translated)
		}
	}
}
