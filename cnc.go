package main

//This is command and control tool

import (
	"compress/gzip"
	"crypto/tls"
	"encoding/gob"
	"encoding/json"
	"errors"
	"flag"
	"io"
	"io/ioutil"
	"log"
	"math/big"
	"net"
	"net/http"
	"net/rpc"
	"strings"
	"sync"
	"time"

	"github.com/miekg/dns"
	"github.com/sajal/mtrparser"
	"github.com/turbobytes/geoipdb"
	"github.com/turbobytes/pulse/utils"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

//type Resolver int
var geo geoipdb.Handler
var session *mgo.Session

//AgentInfo is what we store in db...
type AgentInfo struct {
	Name           string
	City           string
	State          string
	Country        string
	SerialNumber   *big.Int
	LocalResolvers []string
	ASN            string
	ASName         string
	Host           string
	//HostEmail       string
	//HostWebsite     string
	//HostDescription string
	//HostCompanyLogo string
	HostType    string // H = Home, O = Office, D = Datacenter
	FirstOnline string
	LatLng      string //TODO: make richer?
}

func (agent *AgentInfo) GetBSON() (interface{}, error) {
	return bson.D{
		{"Name", agent.Name},
		{"City", agent.City},
		{"State", agent.State},
		{"Country", agent.Country},
		{"LocalResolvers", strings.Join(agent.LocalResolvers, ",")},
		{"_id", agent.SerialNumber.String()},
		{"ASN", agent.ASN},
		{"ASName", agent.ASName},
		{"Host", agent.Host},
		//{"HostWebsite", agent.HostWebsite},
		//{"HostDescription", agent.HostDescription},
		{"HostType", agent.HostType},
		//{"HostCompanyLogo", agent.HostCompanyLogo},
		{"FirstOnline", agent.FirstOnline},
		{"LatLng", agent.LatLng},
	}, nil
}

func (agent *AgentInfo) SetBSON(raw bson.Raw) error {
	data := make(map[string]string)
	err := raw.Unmarshal(data)
	if err != nil {
		return err
	}
	agent.Name = data["Name"]
	agent.City = data["City"]
	agent.State = data["State"]
	agent.Country = data["Country"]
	agent.LocalResolvers = strings.Split(data["LocalResolvers"], ",")
	agent.SerialNumber = new(big.Int)
	agent.SerialNumber.SetString(data["_id"], 10)
	agent.ASN = data["ASN"]
	agent.ASName = data["ASName"]
	agent.Host = data["Host"]
	//agent.HostWebsite = data["HostWebsite"]
	//agent.HostDescription = data["HostDescription"]
	//agent.HostCompanyLogo = data["HostCompanyLogo"]
	agent.HostType = data["HostType"]
	agent.FirstOnline = data["FirstOnline"]
	agent.LatLng = data["LatLng"]
	return nil
}

type Worker struct {
	Client *rpc.Client `json:"date"`
	IP     string      `json:"date"`
	//Geo       string      //TODO: Make richer
	Resolvers []string //List of resolvers this worker supports
	Name      string
	ASN       *string
	ASName    *string
	State     string
	Country   string
	City      string
	Serial    *big.Int
	//HostCompanyLogo string
	//HostWebsite     string
	//HostDescription string
	HostType     string
	Host         string
	LatLng       string //TODO: make richer?
	FirstOnline  string
	connectedat  time.Time
	ConnectedFor string
	Connected    bool
}

func populatedata(w *Worker, insertfirst bool) {
	c := session.DB("dnsdist").C("agents")
	agent := new(AgentInfo)
	err := c.Find(bson.M{"_id": w.Serial.String()}).One(agent)
	if err == mgo.ErrNotFound && insertfirst {
		agent.Name = w.Name
		agent.SerialNumber = w.Serial
		agent.FirstOnline = time.Now().UTC().String()
		err1 := c.Insert(agent)
		if err1 != nil {
			log.Fatal(err1)
		}
	} else if err != nil {
		log.Println(err)
		return
	}
	w.Name = agent.Name
	w.City = agent.City
	w.State = agent.State
	w.Country = agent.Country
	w.Resolvers = agent.LocalResolvers
	w.LatLng = agent.LatLng
	//w.HostDescription = agent.HostDescription
	//w.HostCompanyLogo = agent.HostCompanyLogo
	//w.HostWebsite = agent.HostWebsite
	w.HostType = agent.HostType
	w.Host = agent.Host
	if !insertfirst {
		//Populate is running cause of offline agent
		w.ASN = &agent.ASN
		w.ASName = &agent.ASName
	} else {
		//Update DB with last known ASN data
		c.UpdateId(agent.SerialNumber.String(), bson.M{"$set": bson.M{"ASN": w.ASN, "ASName": w.ASName}})
	}
	if agent.FirstOnline == "" && insertfirst {
		//The first time it actually came online...
		log.Println("This is first time agent came online ", agent.SerialNumber)
		agent.FirstOnline = time.Now().UTC().String()
		c.UpdateId(agent.SerialNumber.String(), bson.M{"$set": bson.M{"FirstOnline": agent.FirstOnline}})
	}
	w.FirstOnline = agent.FirstOnline
}

// lookupAsn is a wrapper around geoipdb.LookupAsn
// that returns results as pointers
func lookupAsn(ip string) (*string, *string) {
	asn, descr, err := geo.LookupAsn(ip)
	if err != nil {
		log.Printf("warning: failed to lookup ASN for %s: %s\n", ip, err)
		return nil, nil
	}
	return &asn, &descr
}

func NewWorker(conn net.Conn) *Worker {
	w := &Worker{}
	w.Client = rpc.NewClient(conn)
	w.IP = strings.Split(conn.RemoteAddr().String(), ":")[0]
	//TODO: Authenticate and fetch capabilities
	w.connectedat = time.Now()
	w.Connected = true
	tlsconn, ok := conn.(*tls.Conn)
	if !ok {
		log.Println("Not TLS Conn")
	} else {
		var err error
		w.ASN, w.ASName = lookupAsn(w.IP)
		err = pingworker(w) //Ping in beginning to make sure we can talk and trigger handshake
		if err == nil {
			state := tlsconn.ConnectionState()
			if len(state.PeerCertificates) > 0 {
				w.Name = state.PeerCertificates[0].Subject.CommonName
				serial := state.PeerCertificates[0].SerialNumber
				log.Println(serial)
				w.Serial = serial
				log.Println(w)
				populatedata(w, true)
				log.Println(w)
				return w
			}
		}
	}
	return nil
}

type Tracker struct {
	workers    map[string]*Worker
	workerlock *sync.RWMutex
}

func NewTracker() *Tracker {
	t := &Tracker{}
	t.workerlock = &sync.RWMutex{}
	t.workers = make(map[string]*Worker)
	go t.Pinger()
	return t
}

func (tracker *Tracker) Register(conn net.Conn) {
	worker := NewWorker(conn)
	if worker != nil {
		tracker.workerlock.Lock()
		tracker.workers[conn.RemoteAddr().String()] = worker
		tracker.workerlock.Unlock()
	}
}

func (tracker *Tracker) UnRegister(worker *Worker) {
	tracker.workerlock.Lock()
	defer tracker.workerlock.Unlock()
	//Copy all except this one
	for k, w := range tracker.workers {
		if worker == w {
			delete(tracker.workers, k)
		}
	}
	//tracker.workers = newworkers
}

func pingworker(worker *Worker) (err error) {
	var reply bool
	c := make(chan error, 1)
	//We use this channel trikery to implement a timeout. If pinger doesn't respond in 10 seconds we kill the connection.
	go func() {
		c <- worker.Client.Call("Pinger.Ping", true, &reply)
	}()
	select {
	case err = <-c:
		if err == rpc.ErrShutdown {
			go tracker.UnRegister(worker) //Async cause of locking
			log.Println("Unregistering from tracker")
		} else if err != nil {
			log.Println("pinger", err)
		}
	case <-time.After(10 * time.Second):
		go tracker.UnRegister(worker) //Did not respond to ping in 10 seconds
		err = errors.New("Ping timeout")
		log.Println(err)
	}
	return err
}

func (tracker *Tracker) SendPings() {
	tracker.workerlock.RLock()
	defer tracker.workerlock.RUnlock()
	for _, worker := range tracker.workers {
		go pingworker(worker)
	}

}

func (tracker *Tracker) Pinger() {
	for {
		time.Sleep(time.Second * 20)
		tracker.SendPings()
	}
}

func addresolvers(args pulse.DNSRequest, resolvers []string) {

}

//Dump json data of all workers...
func (tracker *Tracker) WorkerJson() []byte {
	tracker.workerlock.RLock()
	defer tracker.workerlock.RUnlock()
	workers := make([]*Worker, 0)
	foundids := make([]string, 0)
	for _, w := range tracker.workers {
		w.ConnectedFor = time.Since(w.connectedat).String()
		workers = append(workers, w)
		foundids = append(foundids, w.Serial.String())
	}
	//Append offline workers...
	c := session.DB("dnsdist").C("agents")
	var newids []string
	c.Find(bson.M{"_id": bson.M{"$nin": foundids}}).Distinct("_id", &newids)
	//log.Println(err)
	//log.Println(foundids)
	//log.Println(newids)
	for _, newid := range newids {
		wrk := new(Worker)
		wrk.Serial = new(big.Int)
		wrk.Serial.SetString(newid, 10)
		populatedata(wrk, false)
		workers = append(workers, wrk)
	}
	data, _ := json.MarshalIndent(workers, "", "  ")
	return data
}

func (tracker *Tracker) SingleWorkerJson(agentid string) ([]byte, error) {
	id := new(big.Int)
	id.SetString(agentid, 10)
	tracker.workerlock.RLock()
	defer tracker.workerlock.RUnlock()
	var data []byte
	var wrk *Worker
	for _, w := range tracker.workers {
		w.ConnectedFor = time.Since(w.connectedat).String()
		if id.Cmp(w.Serial) == 0 {
			wrk = w
		}
	}
	//log.Println(wrk)
	if wrk == nil {
		return data, errors.New("Not found")
	}
	data, err := json.MarshalIndent(wrk, "", "  ")
	return data, err
}

//Repopulate the worker info from db... without having to disconnect
func (tracker *Tracker) Repopulate() {
	tracker.workerlock.Lock()
	defer tracker.workerlock.Unlock()
	for _, w := range tracker.workers {
		populatedata(w, true)
	}
}

func slicecontainsbigint(num *big.Int, arr []*big.Int) bool {
	for _, n := range arr {
		if num.Cmp(n) == 0 {
			return true
		}
	}
	return false
}

func (tracker *Tracker) Runner(reqorg *pulse.CombinedRequest) []*pulse.CombinedResult {
	tracker.workerlock.RLock()
	defer tracker.workerlock.RUnlock()
	log.Println(reqorg.AgentFilter)
	var tmpworker = make(map[string]*Worker)
	for ip, worker := range tracker.workers {
		if len(reqorg.AgentFilter) == 0 {
			tmpworker[ip] = worker
		} else if slicecontainsbigint(worker.Serial, reqorg.AgentFilter) {
			tmpworker[ip] = worker
		}
	}
	results := make([]*pulse.CombinedResult, 0)
	n := len(tmpworker)
	rchan := make(chan *pulse.CombinedResult, n)
	var originalargs pulse.DNSRequest
	if reqorg.Type == pulse.TypeDNS {
		args, ok := reqorg.Args.(pulse.DNSRequest)
		if ok {
			originalargs = args
		}
	}
	for ip, worker := range tmpworker {
		go func(worker *Worker, ip string) {
			//Clone the request to avoid pointer mixup when issuing concurrent rpc calls
			req := reqorg.Clone()
			log.Println(ip, worker)
			var reply *pulse.CombinedResult
			//TODO: Implement timeout
			//If CombinedRequest is of type TypeDNS and taget is not specified... then insert defaults for worker...
			if req.Type == pulse.TypeDNS {
				args, ok := req.Args.(pulse.DNSRequest)
				if ok {
					if len(originalargs.Targets) == 0 {
						args.Targets = []string{"8.8.8.8:53", "208.67.222.222:53"}
						for _, resolver := range worker.Resolvers {
							if resolver != "" {
								args.Targets = append(args.Targets, resolver+":53")
							}
						}
						req.Args = args
					}
				}
			}
			call := worker.Client.Go("Resolver.Combined", req, &reply, nil)
			select {
			case replyCall := <-call.Done:
				log.Println(ip)
				if replyCall.Error == rpc.ErrShutdown {
					go tracker.UnRegister(worker) //Async cause of locking
					log.Println("Unregistering from tracker")
					rchan <- nil
				} else if replyCall.Error != nil {
					log.Println(replyCall.Error)
					rchan <- nil
				} else {
					//reply.Name += " (" + strings.Split(ip, ":")[0] + ")"
					iponly := strings.Split(ip, ":")[0]
					splitted := strings.Split(iponly, ".")
					splitted[3] = "0"
					reply.Agent = strings.Join(splitted, ".")
					reply.Name = worker.Name //Insert in this workers Common Name here
					reply.ASN = worker.ASN
					reply.ASName = worker.ASName
					reply.City = worker.City
					reply.State = worker.State
					reply.Country = worker.Country
					reply.Id = worker.Serial
					//log.Println(reply.Name)
					rchan <- reply
				}
				return
			case <-time.After(time.Minute):
				go tracker.UnRegister(worker) //Nuke the turtle...
				rchan <- nil
				return
			}
		}(worker, ip)
	}

	for i := 0; i < n; i++ {
		log.Println(i, "of", n)
		reply := <-rchan
		if reply != nil {
			log.Println(reply.Name)
			results = append(results, reply)
		}
	}
	return results
}

var tracker *Tracker

//https://gist.github.com/the42/1956518

type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func makeGzipHandler(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.Header.Get("Origin"), "https://my.turbobytes.com") {
			w.Header().Set("Access-Control-Allow-Origin", "https://my.turbobytes.com")
			w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
			w.Header().Set("Access-Control-Max-Age", "3600")
			w.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")
		}
		if strings.Contains(r.Header.Get("Origin"), "http://127.0.0.1:8000") {
			w.Header().Set("Access-Control-Allow-Origin", "http://127.0.0.1:8000")
			w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
			w.Header().Set("Access-Control-Max-Age", "3600")
			w.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")
		}
		if r.Method == "OPTIONS" {
			switch {
			default:
				// default OPTIONS method handling
				return
			case strings.Index(r.URL.Path, asndbEndpoint) == 0:
				// Let asndb handler deal with OPTIONS
			case strings.Index(r.URL.Path, asnlookupEndpoint) == 0:
				// Let asnlookup handler deal with OPTIONS
			}
		}
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			fn(w, r)
			return
		}
		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		defer gz.Close()
		gzr := gzipResponseWriter{Writer: gz, ResponseWriter: w}
		fn(gzr, r)
	}
}

func runcurl(w http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println(string(data))
	req := &pulse.CurlRequest{}
	err = json.Unmarshal(data, req)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println(req)
	creq := &pulse.CombinedRequest{
		Type:        pulse.TypeCurl,
		Args:        req,
		RequestedAt: time.Now(),
		AgentFilter: req.AgentFilter,
	}
	results := tracker.Runner(creq)
	b, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		log.Println(err)
		log.Fatal(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}

func agentshandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	splitted := strings.Split(r.URL.Path, "/")
	if len(splitted) == 4 {
		agentid := splitted[2]
		b, err := tracker.SingleWorkerJson(agentid)
		if err != nil {
			log.Println(err)
			wrk := new(Worker)
			wrk.Serial = new(big.Int)
			wrk.Serial.SetString(agentid, 10)
			populatedata(wrk, false)
			if wrk.Name == "" {
				w.WriteHeader(404)
				return
			} else {
				b, err = json.MarshalIndent(wrk, "", "  ")
			}
		}
		if err != nil {
			w.WriteHeader(404)
			return
		} else {
			w.Write(b)
		}
	} else {
		w.Write(tracker.WorkerJson())
	}
}

func getasnmtr(ip string) string {
	asn, _, err := geo.LookupAsn(ip)
	if err != nil {
		log.Printf("warning: asn lookup error for %s: %s\n", ip, err)
		return ""
	}
	return asn
}

func repopulatehandler(w http.ResponseWriter, r *http.Request) {
	//w.Header().Set("Content-Type", "application/json")
	tracker.Repopulate()
	w.Write([]byte("DONE"))
}

func ResolveASNMtr(hop *mtrparser.MtrHop) {
	hop.ASN = make([]string, len(hop.IP))
	for idx, ip := range hop.IP {
		//TODO...
		hop.ASN[idx] = getasnmtr(ip)
	}
}

func runmtr(w http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println(string(data))
	req := pulse.MtrRequest{}
	err = json.Unmarshal(data, &req)
	if err != nil {
		log.Println(err)
		return
	}
	creq := &pulse.CombinedRequest{
		Type:        pulse.TypeMTR,
		Args:        req,
		RequestedAt: time.Now(),
		AgentFilter: req.AgentFilter,
	}
	log.Println(req)
	results := tracker.Runner(creq)
	log.Println("Got results")
	var wg sync.WaitGroup
	for _, res := range results {
		result, _ := res.Result.(pulse.MtrResult)
		if result.Result != nil {
			if result.Err == "" {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for _, hop := range result.Result.Hops {
						ResolveASNMtr(hop)
					}
					result.Result.Summarize(10)
				}()
			}
		}
	}
	wg.Wait()
	log.Println("Populated hostnames")
	b, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		log.Println(err)
		log.Fatal(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}

func runtest(w http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println(string(data))
	req := pulse.DNSRequest{}
	err = json.Unmarshal(data, &req)
	if err != nil {
		log.Println(err)
		return
	}
	if !strings.HasSuffix(req.Host, ".") {
		//Make FQDN
		req.Host = req.Host + "."
	}
	if req.Targets != nil {
		if len(req.Targets) > 0 {
			for i, t := range req.Targets {
				req.Targets[i] = t + ":53"
			}
		}
	}
	creq := &pulse.CombinedRequest{
		Type:        pulse.TypeDNS,
		Args:        req,
		RequestedAt: time.Now(),
		AgentFilter: req.AgentFilter,
	}
	log.Println(req)
	results := tracker.Runner(creq)

	//newresult := make(&pulse.CombinedResult, len(results))

	for i, res := range results {
		result, _ := res.Result.(pulse.DNSResult)
		for j, item := range result.Results {
			item.ASN, item.ASName = lookupAsn(item.Server)
			msg := &dns.Msg{}
			msg.Unpack(item.Raw)
			item.Formated = msg.String()
			item.Msg = msg
			result.Results[j] = item
		}
		results[i].Result = result
	}
	b, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}

// asndbHandler manages the asndb http endpoint
func asndbHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	args := strings.Split(r.URL.Path, "/")
	switch len(args) {
	case 0, 1, 2:
		// url: <nil> or '/' or '/asndb'
		// this should never happen with http.HandleFunc()
		httpInternalServerError(w, errors.New("unexpected asndb url"))
	case 3:
		asn := args[2]
		if asn == "" {
			// url: /asndb/
			allowedMethods := []string{http.MethodOptions, http.MethodGet}
			switch r.Method {
			case http.MethodOptions:
				httpSetAllowHeader(w, allowedMethods)
			case http.MethodGet:
				asndbGet(w)
			default:
				httpMethodNotAllowed(w, allowedMethods)
			}
		} else {
			// url: /asndb/<asn>
			allowedMethods := []string{http.MethodOptions, http.MethodGet, http.MethodPut, http.MethodDelete}
			switch r.Method {
			case http.MethodOptions:
				httpSetAllowHeader(w, allowedMethods)
			case http.MethodGet:
				asndbGetAsn(w, asn)
			case http.MethodPut:
				asndbPutAsn(w, r, asn)
			case http.MethodDelete:
				asndbDeleteAsn(w, asn)
			default:
				httpMethodNotAllowed(w, allowedMethods)
			}
		}
	default:
		// url: /asndb/g/a/r/b/a/g/e
		httpBadRequest(w, errors.New("Too many arguments"))
	}
}

// asndbGet answers all overrides collection.
func asndbGet(w http.ResponseWriter) {
	overrides, err := geo.OverridesList()
	if err != nil {
		if err == geoipdb.OverridesNilCollectionError {
			httpNotAcceptable(w, errors.New("asndb features are disabled"))
			return
		}
		httpInternalServerError(w, err)
		return
	}
	err = httpSendJson(w, overrides)
	if err != nil {
		log.Printf("error: failed to send overrides list: %s", err)
	}
}

// asndbGetAsn retrieves the override description of an ASN.
func asndbGetAsn(w http.ResponseWriter, asn string) {
	descr, err := geo.OverridesLookup(asn)
	if err != nil {
		if err == geoipdb.OverridesNilCollectionError {
			httpNotAcceptable(w, errors.New("asndb features are disabled"))
			return
		}
		if err == geoipdb.OverridesAsnNotFoundError {
			httpNotFound(w)
			return
		}
		httpInternalServerError(w, err)
		return
	}
	err = httpSendJson(w, geoipdb.AsnOverride{Asn: asn, Name: descr})
	if err != nil {
		log.Printf("error: failed to send override value: %s", err)
	}
}

// asndbPutAsn stores an override description of an ASN.
func asndbPutAsn(w http.ResponseWriter, r *http.Request, asn string) {
	cType := r.Header.Get("Content-Type")
	if strings.Index(cType, "application/json") != 0 {
		httpBadRequest(w, errors.New("unexpected content type"))
		return
	}
	var override geoipdb.AsnOverride
	err := json.NewDecoder(r.Body).Decode(&override)
	if err != nil {
		httpBadRequest(w, errors.New("malformed content: "+err.Error()))
		return
	}
	override.Asn = asn
	if override.Name == "" {
		httpBadRequest(w, errors.New("empty name field"))
		return
	}
	err = geo.OverridesSet(override.Asn, override.Name)
	if err != nil {
		if err == geoipdb.OverridesNilCollectionError {
			httpNotAcceptable(w, errors.New("asndb features are disabled"))
			return
		}
		if err == geoipdb.OverridesMalformedAsnError {
			httpBadRequest(w, errors.New("malformed ASN id"))
			return
		}
		httpInternalServerError(w, err)
		return
	}
	httpSendJson(w, override)
}

// types of lookups under /asnlookup/
const (
	asnlookupTypeASN = iota
	asnlookupTypeIP  = iota
)

// asndbDeleteAsn removes the override description of an ASN.
func asndbDeleteAsn(w http.ResponseWriter, asn string) {
	err := geo.OverridesRemove(asn)
	if err != nil {
		if err == geoipdb.OverridesNilCollectionError {
			httpNotAcceptable(w, errors.New("asndb features are disabled"))
			return
		}
		httpInternalServerError(w, err)
		return
	}
}

// asnlookupHandler manages the asnlookup http endpoint
func asnlookupHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	args := strings.Split(r.URL.Path, "/")
	if len(args) < 3 {
		// url: <nil> or '/' or '/asnlookup'
		// this should never happen with http.HandleFunc()
		httpInternalServerError(w, errors.New("unexpected asnlookup url"))
		return
	}
	//url: /asnlookup/[...]
	args = args[2:]
	var lookupType int
	switch args[0] {
	case "":
		// url: /asnlookup/
		httpBadRequest(w, errors.New("missing lookup type"))
		return
	case "asn":
		// url: /asnlookup/asn[/...]
		lookupType = asnlookupTypeASN
	case "ip":
		// url: /asnlookup/ip[/...]
		lookupType = asnlookupTypeIP
	default:
		// url: /asnlookup/_whatever_[/...]
		httpBadRequest(w, errors.New("unexpected lookup type"))
		return
	}
	if len(args) < 2 {
		// url: /asnlookup/<lookupType>
		httpBadRequest(w, errors.New("missing lookup parameter"))
		return
	}
	args = args[1:]
	parameter := args[0]
	if parameter == "" {
		// url: /asnlookup/<lookupType>/
		httpBadRequest(w, errors.New("missing lookup parameter"))
		return
	}
	if len(args) > 1 {
		// url: /asnlookup/<lookupType>/<parameter>/[...]
		httpBadRequest(w, errors.New("too many arguments"))
		return
	}
	// url: /asnlookup/<lookupType>/<parameter>
	allowedMethods := []string{http.MethodOptions, http.MethodGet}
	switch r.Method {
	case http.MethodOptions:
		// OPTIONS /asnlookup/<lookupType>/<parameter>
		httpSetAllowHeader(w, allowedMethods)
		return
	case http.MethodGet:
		// GET /asnlookup/<lookupType>/<parameter>
	default:
		// _OTHER_METHOD_ /asnlookup/<lookupType>/<parameter>
		httpMethodNotAllowed(w, allowedMethods)
		return
	}
	// GET /asnlookup/<lookupType>/<parameter>
	switch lookupType {
	case asnlookupTypeASN:
		asnlookupGetByAsn(w, parameter)
	case asnlookupTypeIP:
		asnlookupGetByIp(w, parameter)
	default:
		httpInternalServerError(w, errors.New("lookup type panic"))
	}
}

// asnlookupResult is answered by /asnlookup/ endpoint.
type AsnlookupResult struct {
	// ASN identification
	Asn string `json:"asn"`
	// IP address
	Ip     string `json:"ip"`
	Result struct {
		// MaxMind GeoIP
		Maxmind AsnlookupQueryResult `json:"maxmind"`
		// ipinfo.io IP lookup API
		Ipinfo AsnlookupQueryResult `json:"ipinfo"`
		// Team Cymru's DNS
		Cymru AsnlookupQueryResult `json:"cymru"`
		// Pulse ASN DB
		Asndb AsnlookupQueryResult `json:"asndb"`
		// TurboBytes geoipdb.LookupAsnP
		Geoipdb AsnlookupQueryResult `json:"geoipdb"`
	} `json:"result"`
}

type AsnlookupQueryResult struct {
	// ASN description
	Name string `json:"name"`
	// status about query
	Err string `json:"err"`
}

// asnlookupGetByAsn queries several sources for ASN descriptions.
// ASN lookup is done by ASN identifier.
func asnlookupGetByAsn(w http.ResponseWriter, asn string) {
	var answer AsnlookupResult
	answer.Asn = asn
	// FIXME: fill IP field. This may be possible after
	//     https://github.com/turbobytes/geoipdb/issues/18
	// Query Cymru
	cymru := make(chan interface{})
	go func () {
		var err error
		answer.Result.Cymru.Name, err = geo.CymruDnsLookup(answer.Asn)
		if err != nil {
			answer.Result.Cymru.Err = err.Error()
		}
		close(cymru)
	}()
	var err error
	answer.Result.Asndb.Name, err = geo.OverridesLookup(answer.Asn)
	if err != nil {
		answer.Result.Asndb.Err = err.Error()
	}
	// Wait for external queries to finish
	<-cymru
	// FIXME: geoipdb lookup
	// FIXME: ipinfo lookup
	// FIXME: maxmind lookup
	httpSendJson(w, answer)
}

// asnlookupGetByIp queries several sources for ASN descriptions.
// ASN lookup is done by IP address.
func asnlookupGetByIp(w http.ResponseWriter, ip string) {
	var answer AsnlookupResult
	answer.Ip = ip
	// Query GeoIPDB, find ASN
	var err error
	answer.Asn, answer.Result.Geoipdb.Name, err = geo.LookupAsn(answer.Ip)
	if err != nil {
		answer.Result.Geoipdb.Err = err.Error()
	}
	// Query Cymru
	cymru := make(chan interface{})
	go func () {
		var err error
		answer.Result.Cymru.Name, err = geo.CymruDnsLookup(answer.Asn)
		if err != nil {
			answer.Result.Cymru.Err = err.Error()
		}
		close(cymru)
	}()
	// Query IPInfo
	ipinfo := make(chan interface{})
	go func() {
		var err error
		_, answer.Result.Ipinfo.Name, err = geo.IpInfoLookup(answer.Ip)
		if err != nil {
			answer.Result.Ipinfo.Err = err.Error()
		}
		close(ipinfo)
	}()
	// Query MaxMind
	_, answer.Result.Maxmind.Name = geo.LibGeoipLookup(answer.Ip)
	// Query AsnDB
	answer.Result.Asndb.Name, err = geo.OverridesLookup(answer.Asn)
	if err != nil {
		answer.Result.Asndb.Err = err.Error()
	}
	// Wait for external queries to finish
	<-cymru
	<-ipinfo
	// Send results
	httpSendJson(w, answer)
}

// httpSendJson sends an object as JSON.
func httpSendJson(w http.ResponseWriter, data interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	return json.NewEncoder(w).Encode(data)
}

// httpBadRequest sends "bad request" http status.
func httpBadRequest(w http.ResponseWriter, err error) {
	http.Error(
		w,
		"Bad Request\n"+err.Error(),
		http.StatusBadRequest,
	)
}

// httpNotFound sends "not found" http status.
func httpNotFound(w http.ResponseWriter) {
	http.Error(w, "Not Found", http.StatusNotFound)
}

// httpMethodNotAllowed sends "method not allowed" http status.
func httpMethodNotAllowed(w http.ResponseWriter, allowed []string) {
	if len(allowed) == 0 {
		httpNotImplemented(w)
		return
	}
	httpSetAllowHeader(w, allowed)
	http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
}

// httpNotAcceptable sends "not acceptable" http status.
func httpNotAcceptable(w http.ResponseWriter, err error) {
	http.Error(
		w,
		"Not Acceptable\n"+err.Error(),
		http.StatusNotAcceptable,
	)
}

// httpInternalServerError sends "internal server error" http status.
func httpInternalServerError(w http.ResponseWriter, err error) {
	http.Error(
		w,
		"Internal Server Error\n"+err.Error(),
		http.StatusInternalServerError,
	)
}

// httpNotImplemented sends "not implemented" http status.
func httpNotImplemented(w http.ResponseWriter) {
	http.Error(w, "Not Implemented", http.StatusNotImplemented)
}

// httpSetAllowHeader sets the "Allow" http response header.
func httpSetAllowHeader(w http.ResponseWriter, allowed []string) {
	w.Header().Set("Allow", strings.Join(allowed, ", "))
}

const (
	asndbEndpoint     = "/asndb/"
	asnlookupEndpoint = "/asnlookup/"
)

func main() {
	gob.RegisterName("github.com/turbobytes/pulse/utils.MtrRequest", pulse.MtrRequest{})
	gob.RegisterName("github.com/turbobytes/pulse/utils.MtrResult", pulse.MtrResult{})
	gob.RegisterName("github.com/turbobytes/pulse/utils.CurlRequest", pulse.CurlRequest{})
	gob.RegisterName("github.com/turbobytes/pulse/utils.CurlResult", pulse.CurlResult{})
	gob.RegisterName("github.com/turbobytes/pulse/utils.DNSRequest", pulse.DNSRequest{})
	gob.RegisterName("github.com/turbobytes/pulse/utils.DNSResult", pulse.DNSResult{})
	tracker = NewTracker()
	var err error
	session, err = mgo.Dial("127.0.0.1")
	if err != nil {
		log.Fatal("mongo ", err)
	}
	defer session.Close()

	geo, err = geoipdb.NewHandler(
		session.DB("dnsdist").C("geoipdb"),
		time.Second*5,
	)
	if err != nil {
		log.Fatalf("failed to get a geoipdb handler: %s", err)
	}

	var caFile, certificateFile, privateKeyFile string
	flag.StringVar(&caFile, "ca", "ca.crt", "Path to CA")
	flag.StringVar(&certificateFile, "crt", "server.crt", "Path to Server Certificate")
	flag.StringVar(&privateKeyFile, "key", "server.key", "Path to Private key")
	flag.Parse()
	cfg := pulse.GetTLSConfig(caFile, certificateFile, privateKeyFile)

	listener, err := tls.Listen("tcp4", ":7777", cfg)
	if err != nil {
		log.Fatal(err)
	}
	go func() {

		http.HandleFunc("/", makeGzipHandler(func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, "index-dist.html")
		}))
		http.HandleFunc("/dns/", makeGzipHandler(runtest))
		http.HandleFunc("/curl/", makeGzipHandler(runcurl))
		http.HandleFunc("/mtr/", makeGzipHandler(runmtr))
		http.HandleFunc("/agents/", makeGzipHandler(agentshandler))
		http.HandleFunc("/repopulate/", makeGzipHandler(repopulatehandler))
		http.HandleFunc(asndbEndpoint, makeGzipHandler(asndbHandler))
		http.HandleFunc(asnlookupEndpoint, makeGzipHandler(asnlookupHandler))

		log.Fatal(http.ListenAndServe(":7778", nil))

	}()
	log.Println("monitoring")
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go tracker.Register(conn) //Async cause this pings also
		log.Println(conn.RemoteAddr(), "at your service")
		//workers[worker.addr.String()] = worker
	}
}
