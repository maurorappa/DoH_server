/*
   DNS-over-HTTPS
   Copyright (C) 2017-2018 Star Brilliant <m13253@hotmail.com>

   Permission is hereby granted, free of charge, to any person obtaining a
   copy of this software and associated documentation files (the "Software"),
   to deal in the Software without restriction, including without limitation
   the rights to use, copy, modify, merge, publish, distribute, sublicense,
   and/or sell copies of the Software, and to permit persons to whom the
   Software is furnished to do so, subject to the following conditions:

   The above copyright notice and this permission notice shall be included in
   all copies or substantial portions of the Software.

   THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
   IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
   FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
   AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
   LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
   FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER
   DEALINGS IN THE SOFTWARE.
*/

package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/ReneKroon/ttlcache"
	"github.com/gorilla/handlers"
	"github.com/m13253/dns-over-https/json-dns"
	"github.com/miekg/dns"
	n "github.com/skarademir/naturalsort"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Server struct {
	conf      *config
	udpClient *dns.Client
	tcpClient *dns.Client
	servemux  *http.ServeMux
}

type DNSRequest struct {
	request         *dns.Msg
	response        *dns.Msg
	transactionID   uint16
	currentUpstream string
	isTailored      bool
	errcode         int
	errtext         string
}

var (
	rtimes     []string
	query_loop int = 0
	dnscache   *ttlcache.Cache
)

func init() {
	dnscache = ttlcache.NewCache()
	defer dnscache.Close()
}

func NewServer(conf *config) (s *Server) {
	s = &Server{
		conf: conf,
		udpClient: &dns.Client{
			Net:     "udp",
			UDPSize: dns.DefaultMsgSize,
			Timeout: time.Duration(conf.Timeout) * time.Second,
		},
		tcpClient: &dns.Client{
			Net:     "tcp",
			Timeout: time.Duration(conf.Timeout) * time.Second,
		},
		servemux: http.NewServeMux(),
	}

	s.servemux.HandleFunc("/stat", s.handlerFuncStat)
	s.servemux.HandleFunc(conf.Path, s.handlerFunc)
	fmt.Printf("Listening on %s...\n", conf.Listen)
	return
}

func (s *Server) Start() error {
	servemux := http.Handler(s.servemux)
	if s.conf.Verbose {
		servemux = handlers.CombinedLoggingHandler(os.Stdout, servemux)
	}
	results := make(chan error, len(s.conf.Listen))
	for _, addr := range s.conf.Listen {
		go func(addr string) {
			var err error
			certFile, err := ioutil.ReadFile(s.conf.Cert)
			if err != nil {
				log.Fatal(err)
			}

			block, _ := pem.Decode(certFile)
			if block == nil {
				log.Fatal("failed to decode PEM block containing public key")
			}

			// trim the bytes to actual length in call
			cert, err := x509.ParseCertificate(block.Bytes)
			if err != nil {
				log.Fatal(err)
			}

			fmt.Printf("Certificate CN: %s\n", cert.Subject.CommonName)
			fmt.Printf("Validity: Not before %s and ", cert.NotBefore.String())
			fmt.Printf("Not after %s\n", cert.NotAfter.String())

			cfg := &tls.Config{
				MinVersion:               tls.VersionTLS12,
				CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
				PreferServerCipherSuites: true,
				CipherSuites: []uint16{
					tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
					tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
					tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
					tls.TLS_RSA_WITH_AES_256_CBC_SHA,
				},
			}
			srv := &http.Server{
				Addr:         addr,
				Handler:      servemux,
				TLSConfig:    cfg,
				TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0),
			}
			err = srv.ListenAndServeTLS(s.conf.Cert, s.conf.Key)
			if err != nil {
				log.Println(err)
			}
			results <- err
		}(addr)
	}
	// wait for all handlers
	for i := 0; i < cap(results); i++ {
		err := <-results
		if err != nil {
			return err
		}
	}
	close(results)
	return nil
}

func (s *Server) handlerFuncStat(w http.ResponseWriter, r *http.Request) {
	reply := "Resolver Stats\n"
	sort.Sort(n.NaturalSort(rtimes))
	for _, v := range rtimes {
		reply = reply + v + ",\n"
	}
	reply = reply + "\n\n"
	for k, v := range dns_stat {
		reply = reply + k + ": " + strconv.Itoa(v) + ", "
	}
	count := dnscache.Count()
	reply = reply + "\n\nCached entries: " + strconv.Itoa(count)
	w.Write([]byte(reply))
}

func (s *Server) handlerFunc(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Server", "DoH server")

	if r.Form == nil {
		const maxMemory = 32 << 20 // 32 MB
		r.ParseMultipartForm(maxMemory)
	}
	contentType := r.Header.Get("Content-Type")
	if ct := r.FormValue("ct"); ct != "" {
		contentType = ct
	}
	if contentType == "" {
		// Guess request Content-Type based on other parameters
		if r.FormValue("name") != "" {
			contentType = "application/dns-json"
		} else if r.FormValue("dns") != "" {
			contentType = "application/dns-message"
		}
	}
	var responseType string
	for _, responseCandidate := range strings.Split(r.Header.Get("Accept"), ",") {
		responseCandidate = strings.SplitN(responseCandidate, ";", 2)[0]
		if responseCandidate == "application/json" {
			responseType = "application/json"
			break
		} else if responseCandidate == "application/dns-udpwireformat" {
			responseType = "application/dns-message"
			break
		} else if responseCandidate == "application/dns-message" {
			responseType = "application/dns-message"
			break
		}
	}
	if responseType == "" {
		// Guess response Content-Type based on request Content-Type
		if contentType == "application/dns-json" {
			responseType = "application/json"
		} else if contentType == "application/dns-message" {
			responseType = "application/dns-message"
		} else if contentType == "application/dns-udpwireformat" {
			responseType = "application/dns-message"
		}
	}

	var req *DNSRequest
	var err error
	if contentType == "application/dns-json" {
		req = s.parseRequestGoogle(w, r)
	} else if contentType == "application/dns-message" {
		req = s.parseRequestIETF(w, r)
	} else if contentType == "application/dns-udpwireformat" {
		req = s.parseRequestIETF(w, r)
	} else {
		jsonDNS.FormatError(w, fmt.Sprintf("Invalid argument value: \"ct\" = %q", contentType), 415)
		return
	}
	if req.errcode == 444 {
		return
	}
	if req.errcode != 0 {
		jsonDNS.FormatError(w, req.errtext, req.errcode)
		return
	}

	req = s.patchRootRD(req)
	//fmt.Printf("Asking for :%s\n",req.request)

	//check if we have a cached result
	dnsResult, exists := dnscache.Get(req.request.String())
	if !exists {
		dnsResult, err = s.doDNSQuery(req)
		if err != nil {
			jsonDNS.FormatError(w, fmt.Sprintf("DNS query failure (%s)", err.Error()), 503)
			return
		}
		dnscache.SetWithTTL(req.request.String(), dnsResult, 60*time.Second)
	}

	if responseType == "application/json" {
		s.generateResponseGoogle(w, r, req)
	} else if responseType == "application/dns-message" {
		s.generateResponseIETF(w, r, req)
	} else {
		panic("Unknown response Content-Type")
	}
}

func (s *Server) findClientIP(r *http.Request) net.IP {
	XForwardedFor := r.Header.Get("X-Forwarded-For")
	if XForwardedFor != "" {
		for _, addr := range strings.Split(XForwardedFor, ",") {
			addr = strings.TrimSpace(addr)
			ip := net.ParseIP(addr)
			if jsonDNS.IsGlobalIP(ip) {
				return ip
			}
		}
	}
	XRealIP := r.Header.Get("X-Real-IP")
	if XRealIP != "" {
		addr := strings.TrimSpace(XRealIP)
		ip := net.ParseIP(addr)
		if jsonDNS.IsGlobalIP(ip) {
			return ip
		}
	}
	remoteAddr, err := net.ResolveTCPAddr("tcp", r.RemoteAddr)
	if err != nil {
		return nil
	}
	if ip := remoteAddr.IP; jsonDNS.IsGlobalIP(ip) {
		return ip
	}
	return nil
}

// Workaround a bug causing Unbound to refuse returning anything about the root
func (s *Server) patchRootRD(req *DNSRequest) *DNSRequest {
	for _, question := range req.request.Question {
		if question.Name == "." {
			req.request.RecursionDesired = true
		}
	}
	return req
}

func (s *Server) doDNSQuery(req *DNSRequest) (resp *DNSRequest, err error) {
	numServers := len(s.conf.Upstream)
	rtt := time.Duration(0)
	if s.conf.Verbose {
		fmt.Printf("\tQuery number: %d\n", query_loop)
	}
	if query_loop <= 10 { // first 10 times I pick random dns
		req.currentUpstream = s.conf.Upstream[rand.Intn(numServers)]
	} else if query_loop > 10 && query_loop < 100 { // I use the fastest
		sort.Sort(n.NaturalSort(rtimes))
		dns := strings.Split(rtimes[0], "-")
		req.currentUpstream = dns[1] //fastest
	} else { // after 90 times, I reset and use random
		query_loop = 0
		req.currentUpstream = s.conf.Upstream[rand.Intn(numServers)]
	}

	for i := uint(0); i < s.conf.Tries; i++ {
		if !s.conf.TCPOnly {
			req.response, rtt, err = s.udpClient.Exchange(req.request, req.currentUpstream)
			if err != nil {
				log.Println(err)
				req.response, rtt, err = s.tcpClient.Exchange(req.request, req.currentUpstream)
			}
		} else {
			req.response, rtt, err = s.tcpClient.Exchange(req.request, req.currentUpstream)
		}
		if s.conf.Verbose {
			log.Printf("DNS server %s, request duration %s", req.currentUpstream, rtt.String())
		}
		//replace(req.currentUpstream, int(rtt))
		replace(req.currentUpstream, strings.Split(string(rtt.String()), ".")[0])
		query_loop = query_loop + 1
		dns_stat[req.currentUpstream]++
		//for _, value := range rtimes {
		//	fmt.Println(value)
		//}
		if err == nil {
			return req, nil
		}
		log.Printf("DNS error from upstream %s: %s\n", req.currentUpstream, err.Error())
	}
	return req, err
}

func replace(dns string, rtt string) {
	found := false
	for k, value := range rtimes {
		if strings.Contains(value, dns) {
			dns := strings.Split(value, "-")
			//fmt.Println(dns[1])
			//fmt.Printf("found it,  %d %s\n",k, value)
			rtimes[k] = rtt + "-" + dns[1]
			found = true
		}
	}
	if !found {
		rtimes = append(rtimes, rtt+"-"+dns)
	}
}
