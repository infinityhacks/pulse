package pulse

import (
	"errors"
	"net"

	"github.com/turbobytes/geoipdb/iputils"
)

var (
	securityerr              = errors.New("Security error: Not allowed to connect to local IP")
	tlsHandshakeTimeoutError = errors.New("net/http: TLS handshake timeout")
)

var (
	// If not nil, these override CIDR tables of islocalip.
	localipv4 []string
	localipv6 []string
)

// islocalip is a wrapper around geoipdb/iputils.IsLocalIP
// that allows tuning CIDR tables in order to mock results.
func islocalip(ip net.IP) bool {
	if localipv4 == nil && localipv6 == nil {
		return iputils.IsLocalIP(ip)
	}
	ipv4 := ip.To4()
	if ipv4 != nil {
		if localipv4 != nil {
			for _, cidr := range localipv4 {
				_, inet, _ := net.ParseCIDR(cidr)
				if inet.Contains(ipv4) {
					return true
				}
			}
			return false
		} else {
			return iputils.IsLocalIP(ipv4)
		}
	}
	ipv6 := ip.To16()
	if ipv6 != nil {
		if localipv6 != nil {
			for _, cidr := range localipv6 {
				_, inet, _ := net.ParseCIDR(cidr)
				if inet.Contains(ipv6) {
					return true
				}
			}
			return false
		} else {
			return iputils.IsLocalIP(ipv6)
		}
	}
	return false
}
