package pulse

import (
	"net"
	"testing"
)

func TestPublicIP(t *testing.T) {
	cases_public := []string{"8.8.8.8", "74.125.130.100", "1.1.1.1", "45.45.45.45", "120.222.111.222", "2404:6800:4003:c01::64"}
	for _, ipstr := range cases_public {
		if islocalip(net.ParseIP(ipstr)) {
			t.Error("Should be false for " + ipstr)
		}
	}
}

func TestLocalIP(t *testing.T) {
	cases_private := []string{"127.0.0.1", "10.5.6.4", "192.168.5.99", "100.66.55.66", "fd07:a47c:3742:823e:3b02:76:982b:463", "::1"}
	for _, ipstr := range cases_private {
		if !islocalip(net.ParseIP(ipstr)) {
			t.Error("Should be true for " + ipstr)
		}
	}
}

//TestOverrideSecurity demonstrates how to override security checks for testing purposes
func TestOverrideSecurity(t *testing.T) {
	originallocalipv4 := localipv4
	localipv4 = []string{}
	originallocalipv6 := localipv6
	localipv6 = []string{}
	//None of these should be security issue
	cases_private := []string{"127.0.0.1", "10.5.6.4", "192.168.5.99", "100.66.55.66", "fd07:a47c:3742:823e:3b02:76:982b:463", "::1"}
	for _, ipstr := range cases_private {
		if islocalip(net.ParseIP(ipstr)) {
			t.Errorf("Should be false for %s because we blanked it", ipstr)
		}
	}
	//Restore original behaviour
	localipv4 = originallocalipv4
	localipv6 = originallocalipv6
	//Test again to see if we could restore it
	for _, ipstr := range cases_private {
		if !islocalip(net.ParseIP(ipstr)) {
			t.Error("Should be true for " + ipstr)
		}
	}
}
