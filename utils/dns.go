package pulse

import (
	"context"
	"log"
	"math/big"
	"strings"
	"time"

	"github.com/miekg/dns"
)

// DNS client timeouts
var (
	dnsDialTimeout  = time.Second * 2
	dnsReadTimeout  = time.Second * 2
	dnsWriteTimeout = time.Second * 2
)

type IndividualDNSResult struct {
	Server     string        //IP/hostname the query was sent to
	Err        string        //Any error that occurred with this particular query.
	ErrEnglish string        //Human friendly version of Err
	RttStr     string        //Round trip time in humanized form
	Rtt        time.Duration //Round trip time
	Raw        []byte        //Raw packet
	Formated   string        //Dig style formating
	Msg        *dns.Msg      //Parsed DNS message
	ASN        *string       //ASN of Server
	ASName     *string       //ASN description
}

type DNSResult struct {
	Results []IndividualDNSResult
	Err     string //Error with this test
}

type DNSRequest struct {
	Host        string   //The DNS query
	QType       uint16   //Query type : https://en.wikipedia.org/wiki/List_of_DNS_record_types#Resource_records
	Targets     []string //The target nameservers
	NoRecursion bool     //true means RecursionDesired = false. false means RecursionDesired = true
	AgentFilter []*big.Int
}

func rundnsquery(host, server string, ch chan IndividualDNSResult, qclass uint16, norecurse, retry bool) {
	res := IndividualDNSResult{}
	res.Server = strings.Split(server, ":")[0]
	m1 := new(dns.Msg)
	m1.Id = dns.Id()
	m1.RecursionDesired = !norecurse
	m1.Question = make([]dns.Question, 1)
	m1.Question[0] = dns.Question{host, qclass, dns.ClassINET}
	c := new(dns.Client)
	c.DialTimeout = dnsDialTimeout
	c.ReadTimeout = dnsReadTimeout
	c.WriteTimeout = dnsWriteTimeout
	log.Println("Asking", server, "for", host)
	msg, rtt, err := c.Exchange(m1, server)
	res.RttStr = rtt.String()
	res.Rtt = rtt
	if err != nil {
		res.Err = err.Error()
		if retry {
			//If fail at first... try again .. once...
			//I could tell a UDP joke... but you might not get it...
			rundnsquery(host, server, ch, qclass, norecurse, false)
		} else {
			ch <- res
		}
	} else {
		//res.Result = msg.String()
		res.Raw, _ = msg.Pack()
		//res.Formated = msg.String()
		ch <- res
	}
}

func rundnsqueryCtx(ctx context.Context, host, server string, ch chan IndividualDNSResult, qclass uint16, norecurse, retry bool) {
	ctxCh := make(chan IndividualDNSResult)
	go rundnsquery(host, server, ctxCh, qclass, norecurse, retry)
	select {
	case res := <-ctxCh:
		ch <- res
	case <-ctx.Done():
		ch <- IndividualDNSResult{
			Server: strings.Split(server, ":")[0],
			Err:    "context deadline exceeded",
		}
	}
}

func DNSImpl(ctx context.Context, r *DNSRequest) *DNSResult {
	//TODO: validate r.Target before sending
	res := new(DNSResult)
	n := len(r.Targets)
	res.Results = make([]IndividualDNSResult, n)
	ch := make(chan IndividualDNSResult, n)
	for _, server := range r.Targets {
		go rundnsqueryCtx(ctx, r.Host, server, ch, r.QType, r.NoRecursion, true)
		time.Sleep(time.Millisecond * 5) //Pace out the packets a bit
	}
	for i := 0; i < n; i++ {
		item := <-ch
		translateDnsError(&item)
		res.Results[i] = item
		//res := runquery(*host, server)
	}

	return res
}
