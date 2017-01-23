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

	"github.com/miekg/dns"
)

func TestTranslateErrorStatic(t *testing.T) {
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
					//DNSTime:1.32406652e+08,
					//DialTime:-9.22337203672237e+18,
					//TLSTime:0,
					//Ttfb:0,
					Err: "Get http://p.catchpoint.com/: dial tcp: lookup p.catchpoint.com on 192.168.1.1:53: no such host",
				},
			},
			"DNS lookup failed. p.catchpoint.com could not be resolved (NXDOMAIN).",
		},
		testCase{
			CombinedResult{
				Type: TypeCurl,
				Result: &CurlResult{
					ConnectTime: 1.5001603612e+10,
					DNSTime:     0,
					DialTime:    1.5001603612e+10,
					TLSTime:     0,
					Ttfb:        0,
					Err:         "Get http://8.8.8.8/: dial tcp 8.8.8.8:80: i/o timeout",
				},
			},
			"Connection timed out. Could not connect to 8.8.8.8:80 within 15 seconds.",
		},
		testCase{
			CombinedResult{
				Type: TypeCurl,
				Result: &CurlResult{
					ConnectTime: 1.5001603612e+10,
					DNSTime:     0,
					DialTime:    1.5001603612e+10,
					TLSTime:     0,
					Ttfb:        0,
					Err:         "Get http://2400:cb00:2048:1::c629:d7a2/: dial tcp [2400:cb00:2048:1::c629:d7a2]:80: i/o timeout",
				},
			},
			"Connection timed out. Could not connect to [2400:cb00:2048:1::c629:d7a2]:80 within 15 seconds.",
		},
		testCase{
			CombinedResult{
				Type: TypeCurl,
				Result: &CurlResult{
					ConnectTime: 0,
					DNSTime:     1.5001603612e+10,
					DialTime:    1.5001603612e+10,
					TLSTime:     0,
					Ttfb:        0,
					Err:         "Get http://some.site.com/: dial tcp some.site.com:80: i/o timeout",
				},
			},
			"DNS lookup timed out. Could not resolve some.site.com within 15 seconds.",
		},
		testCase{
			CombinedResult{
				Type: TypeCurl,
				Result: &CurlResult{
					ConnectTime: 1.4101603612e+10,
					DNSTime:     0.1001603612e+10,
					DialTime:    1.5001603612e+10,
					TLSTime:     0,
					Ttfb:        0,
					Err:         "Get http://some.site.com/: dial tcp some.site.com:80: i/o timeout",
				},
			},
			"Could not connect to some.site.com:80 within 14 seconds. (DNS lookup 1002ms)",
		},
		testCase{
			CombinedResult{
				Type: TypeCurl,
				Result: &CurlResult{
					ConnectTime: 0.9001603612e+10,
					DNSTime:     0.6001603612e+10,
					DialTime:    1.5001603612e+10,
					TLSTime:     0,
					Ttfb:        0,
					Err:         "Get http://some.site.com/: dial tcp some.site.com:80: i/o timeout",
				},
			},
			"Lookup with connection timed out. Could not perform DNS lookup and TCP connection to some.site.com within 15 seconds. (DNS lookup 6002ms, TCP connect 9002ms)",
		},
		testCase{
			CombinedResult{
				Type: TypeCurl,
				Result: &CurlResult{
					ConnectTime: 9.3473233e+07,
					DNSTime:     8.6351441e+07,
					DialTime:    1.79824674e+08,
					TLSTime:     110043,
					Ttfb:        0,
					Err:         "Get http://some.site.com/1234/: net/http: timeout awaiting response headers",
				},
			},
			"Request timed out. TCP connection was established but server did not respond to the request within 25 seconds. (DNS lookup 86ms, TCP connect 93ms, TLS handshake 0ms)",
		},
		testCase{
			CombinedResult{
				Type: TypeCurl,
				Result: &CurlResult{
					//ConnectTime:9.223372036854776e+18,
					DNSTime: 5.01085708e+09,
					//DialTime:-9.223372031843919e+18,
					TLSTime: 0,
					Ttfb:    0,
					Err:     "Get http://lw.cdnplanet.com/static/rum/15kb-image.jpg?t=foo: dial tcp: lookup lw.cdnplanet.com on 8.8.4.4:53: dial udp 8.8.4.4:53: i/o timeout",
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
		testCase{
			CombinedResult{
				Type: TypeCurl,
				Result: &CurlResult{
					Err: "Get https://prod.www-fastly-com.map.fastlylb.net./: x509: certificate is valid for a.ssl.fastly.net, *.a.ssl.fastly.net, rvm.io, not www.nos.nl",
				},
			},
			"Certificate is not valid for www.nos.nl",
		},
		testCase{
			CombinedResult{
				Type: TypeCurl,
				Result: &CurlResult{
					Err: "Get https://some.site.com/ali-mod/??alicloud-assets-footer/0.0.86/index.css: net/http: TLS handshake timeout",
				},
			},
			"TLS handshake timed out.",
		},
		testCase{
			CombinedResult{
				Type: TypeDNS,
				Result: &DNSResult{
					Results: []IndividualDNSResult{
						IndividualDNSResult{
							Err: "dial udp: lookup some.site.com on 192.168.2.254:53: no such host",
						},
					},
				},
			},
			"DNS lookup failed. some.site.com could not be resolved (NXDOMAIN).",
		},
		testCase{
			CombinedResult{
				Type: TypeDNS,
				Result: &DNSResult{
					Results: []IndividualDNSResult{
						IndividualDNSResult{
							Rtt: 2.000831501e+09,
							Err: "read udp 192.168.0.13:55155->208.97.182.10:53: i/o timeout",
						},
					},
				},
			},
			"DNS lookup timed out. No response from 208.97.182.10:53 within 2 seconds.",
		},
		testCase{
			CombinedResult{
				Type: TypeDNS,
				Result: &DNSResult{
					Results: []IndividualDNSResult{
						IndividualDNSResult{
							Err: "read udp 83.169.184.99:53: connection refused",
						},
					},
				},
			},
			"DNS lookup refused. 83.169.184.99 refused to accept the DNS query on port 53. Maybe nothing is listening on that port or a firewall is blocking.",
		},
		testCase{
			CombinedResult{
				Type: TypeDNS,
				Result: &DNSResult{
					Results: []IndividualDNSResult{
						IndividualDNSResult{
							Err: "read udp [2400:cb00:2048:1::c629:d7a2]:53: connection refused",
						},
					},
				},
			},
			"DNS lookup refused. 2400:cb00:2048:1::c629:d7a2 refused to accept the DNS query on port 53. Maybe nothing is listening on that port or a firewall is blocking.",
		},
		testCase{
			CombinedResult{
				Type: TypeDNS,
				Result: &DNSResult{
					Results: []IndividualDNSResult{
						IndividualDNSResult{
							Rtt:    0,
							Err:    "dial udp: i/o timeout",
							Server: "name.server.com",
						},
					},
				},
			},
			"DNS lookup timed out. Could not resolve name.server.com to an IP address within 5 seconds.",
		},
	}
	for _, testCase := range testCases {
		translateError(&testCase.testResult)
		var testType string
		var translated string
		switch testCase.testResult.Type {
		case TypeCurl:
			testType = "HTTP"
			translated = testCase.testResult.Result.(*CurlResult).ErrEnglish
		case TypeDNS:
			testType = "DNS"
			translated = testCase.testResult.Result.(*DNSResult).Results[0].ErrEnglish
		}
		if translated != testCase.expected {
			t.Fatalf("%s error translation mismatch: expected \"%s\", got \"%s\"", testType, testCase.expected, translated)
		}
	}
}

func TestTranslateErrorCurl(t *testing.T) {
	type testCase struct {
		request  *CurlRequest
		expected string
	}
	testCases := []testCase{
		testCase{
			&CurlRequest{
				Path:     "/",
				Endpoint: "some.site.com",
			},
			"DNS lookup failed. some.site.com could not be resolved (NXDOMAIN).",
		},
		testCase{
			&CurlRequest{
				Path:     "/",
				Endpoint: "8.8.8.8",
			},
			"Connection timed out. Could not connect to 8.8.8.8:80 within 15 seconds.",
		},
	}
	for _, testCase := range testCases {
		resp := CurlImpl(testCase.request)
		if resp.Err != "" && resp.ErrEnglish != testCase.expected {
			t.Log(resp)
			t.Errorf("%s error translation mismatch for error '%s': expected \"%s\", got \"%s\"", "HTTP", resp.Err, testCase.expected, resp.ErrEnglish)
		}
	}
}

func TestTranslateErrorDns(t *testing.T) {
	type testCase struct {
		request  *DNSRequest
		expected string
	}
	testCases := []testCase{
		testCase{
			&DNSRequest{
				Host:        "some.site.com.",
				QType:       dns.TypeA,
				Targets:     []string{"unresolvable.nameserver:53"},
				NoRecursion: false,
			},
			"DNS lookup failed. unresolvable.nameserver could not be resolved (NXDOMAIN).",
		},
	}
	for _, testCase := range testCases {
		resp := DNSImpl(testCase.request)
		translated := resp.Results[0].ErrEnglish
		if resp.Results[0].Err != "" && translated != testCase.expected {
			t.Log(resp.Results[0])
			t.Errorf("%s error translation mismatch: expected \"%s\", got \"%s\"", "DNS", testCase.expected, translated)
		}
	}
}
