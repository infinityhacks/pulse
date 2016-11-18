package pulse

import (
	"github.com/turbobytes/geoipdb"
)

// LookupAsn is a wrapper around geoipdb.LookupAsn
// that returns results as pointers
func LookupAsn(h geoipdb.Handler, ip string) (*string, *string, error) {
	asn, descr, err := h.LookupAsn(ip)
	return &asn, &descr, err
}
