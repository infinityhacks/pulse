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
)

func TestTranslateError(t *testing.T) {
	type testCase struct {
		testResult CombinedResult
		expected   string
	}
	testCases := []testCase{
		testCase{
			CombinedResult{
				Type: TypeCurl,
				Result: &CurlResult{
					//ConnectTime: 9.223372036854776e+18,
					//ConnectTimeStr: "2562047h47m16.854775807s",
					//DNSTime:1.32406652e+08,
					//DNSTimeStr:"132.406652ms",
					//DialTime:-9.22337203672237e+18,
					//DialTimeStr:"-2562047h47m16.722369157s",
					//TLSTime:0,
					//TLSTimeStr:"0s",
					//Ttfb:0,
					//TtfbStr:"0s",
					Err: "Get http://p.catchpoint.com/: dial tcp: lookup p.catchpoint.com on 192.168.1.1:53: no such host",
				},
			},
			"DNS lookup failed. p.catchpoint.com could not be resolved (NXDOMAIN).",
		},
		testCase{
			CombinedResult{
				Type: TypeCurl,
				Result: &CurlResult{
					ConnectTime:    1.5001603612e+10,
					ConnectTimeStr: "15.001603612s",
					DNSTime:        0,
					DNSTimeStr:     "0s",
					DialTime:       1.5001603612e+10,
					DialTimeStr:    "15.001603612s",
					TLSTime:        0,
					TLSTimeStr:     "0s",
					Ttfb:           0,
					TtfbStr:        "0s",
					Err:            "Get http://8.8.8.8/: dial tcp 8.8.8.8:80: i/o timeout",
				},
			},
			"Connection timed out. Agent/client could not connect to 8.8.8.8:80 within 15 seconds. (DNS lookup 0s, TCP connect 15.001603612s)",
		},
		testCase{
			CombinedResult{
				Type: TypeCurl,
				Result: &CurlResult{
					ConnectTime:    9.3473233e+07,
					ConnectTimeStr: "93.473233ms",
					DNSTime:        8.6351441e+07,
					DNSTimeStr:     "86.351441ms",
					DialTime:       1.79824674e+08,
					DialTimeStr:    "179.824674ms",
					TLSTime:        110043,
					TLSTimeStr:     "110.043Âµs",
					Ttfb:           0,
					TtfbStr:        "0s",
					Err:            "Get http://some.site.com/1234/: net/http: timeout awaiting response headers",
				},
			},
			"Request timed out. TCP connection was established but server did not respond to the request within 25 seconds. (DNS lookup 86.351441ms, TCP connect 93.473233ms, TLS handshake 110.043µs)",
		},
		testCase{
			CombinedResult{
				Type: TypeCurl,
				Result: &CurlResult{
					//ConnectTime:9.223372036854776e+18,
					//ConnectTimeStr:"2562047h47m16.854775807s",
					DNSTime:    5.01085708e+09,
					DNSTimeStr: "5.01085708s",
					//DialTime:-9.223372031843919e+18,
					//DialTimeStr:"-2562047h47m11.843918729s",
					TLSTime:    0,
					TLSTimeStr: "0s",
					Ttfb:       0,
					TtfbStr:    "0s",
					Err:        "Get http://lw.cdnplanet.com/static/rum/15kb-image.jpg?t=foo: dial tcp: lookup lw.cdnplanet.com on 8.8.4.4:53: dial udp 8.8.4.4:53: i/o timeout",
				},
			},
			"DNS lookup timed out. No response from 8.8.4.4:53 within 5 seconds.",
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
