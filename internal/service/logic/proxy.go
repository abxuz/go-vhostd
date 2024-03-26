package logic

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/abxuz/b-tools/bset"
	"github.com/abxuz/b-tools/bstate"
	"github.com/abxuz/go-vhostd/assets"
	"github.com/abxuz/go-vhostd/internal/model"
	"github.com/abxuz/go-vhostd/internal/service"
	"github.com/abxuz/go-vhostd/utils"
	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
	"golang.org/x/crypto/ocsp"
	"golang.org/x/sync/errgroup"
)

var (
	ErrVhostNotFound = errors.New("vhost not found")
	ErrCertNotFound  = errors.New("cert not found")

	HttpTransport = &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
	}
	Http3Transport = &http3.RoundTripper{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		EnableDatagrams: true,
	}
)

type Mapping struct {
	model.MappingCfg
	Target    *url.URL
	AddHeader http.Header
}

type GetCertificateFunc = func(chi *tls.ClientHelloInfo) (*tls.Certificate, error)

type lProxy struct {
	state *bstate.State[model.Cfg]

	getHttpsCertificate GetCertificateFunc
	getHttp3Certificate GetCertificateFunc

	httpHandler  http.Handler
	httpsHandler http.Handler
	http3Handler http.Handler

	httpServers  map[string]*http.Server
	httpsServers map[string]*http.Server
	http3Servers map[string]*http3.Server
}

func init() {
	service.RegisterProxyService(&lProxy{})
}

func (l *lProxy) Init() {
	l.state = bstate.NewState[model.Cfg]()

	var (
		certs           = make(map[string]*tls.Certificate)
		httpsCerts      = make(map[string]*tls.Certificate)
		http3Certs      = make(map[string]*tls.Certificate)
		certsUpdateLock = new(sync.RWMutex)
	)

	l.state.Watch("Proxy.UpdateCerts", func(_, cfg model.Cfg) {
		certsUpdateLock.Lock()
		defer certsUpdateLock.Unlock()

		clear(certs)
		clear(httpsCerts)
		clear(http3Certs)

		for _, c := range cfg.Cert {
			certs[c.Name], _ = c.Certificate()
		}
		for _, v := range cfg.Https.Vhost {
			httpsCerts[v.Domain] = certs[v.Cert]
		}
		for _, v := range cfg.Http3.Vhost {
			http3Certs[v.Domain] = certs[v.Cert]
		}
	})

	l.getHttpsCertificate = l.newGetCertificateFunc(certsUpdateLock, httpsCerts)
	l.getHttp3Certificate = l.newGetCertificateFunc(certsUpdateLock, http3Certs)
	go l.timerUpdateOCSP(certsUpdateLock, certs)

	var (
		httpLock   = new(sync.RWMutex)
		httpVhost  = make(map[string][]*Mapping)
		httpsLock  = new(sync.RWMutex)
		httpsVhost = make(map[string][]*Mapping)
		http3Lock  = new(sync.RWMutex)
		http3Vhost = make(map[string][]*Mapping)
	)

	l.state.Watch("Proxy.UpdateVhost", func(_, cfg model.Cfg) {
		httpLock.Lock()
		clear(httpVhost)
		for _, vhost := range cfg.Http.Vhost {
			mappings := make([]*Mapping, 0)
			for _, m := range vhost.Mapping {
				mapping := &Mapping{}
				mapping.MappingCfg = *m
				mapping.Target, _ = m.GetTarget()
				mapping.AddHeader, _ = m.GetAddHeader()
				mappings = append(mappings, mapping)
			}
			httpVhost[vhost.Domain] = mappings
		}
		httpLock.Unlock()

		httpsLock.Lock()
		clear(httpsVhost)
		for _, vhost := range cfg.Https.Vhost {
			mappings := make([]*Mapping, 0)
			for _, m := range vhost.Mapping {
				mapping := &Mapping{}
				mapping.MappingCfg = *m
				mapping.Target, _ = m.GetTarget()
				mapping.AddHeader, _ = m.GetAddHeader()
				mappings = append(mappings, mapping)
			}
			httpsVhost[vhost.Domain] = mappings
		}
		httpsLock.Unlock()

		http3Lock.Lock()
		clear(http3Vhost)
		for _, vhost := range cfg.Http3.Vhost {
			mappings := make([]*Mapping, 0)
			for _, m := range vhost.Mapping {
				mapping := &Mapping{}
				mapping.MappingCfg = *m
				mapping.Target, _ = m.GetTarget()
				mapping.AddHeader, _ = m.GetAddHeader()
				mappings = append(mappings, mapping)
			}
			http3Vhost[vhost.Domain] = mappings
		}
		http3Lock.Unlock()

	})

	l.httpHandler = l.newReverseProxy(httpLock, httpVhost)
	l.httpsHandler = l.newReverseProxy(httpsLock, httpsVhost)
	l.http3Handler = l.newReverseProxy(http3Lock, http3Vhost)

	l.httpServers = make(map[string]*http.Server)
	l.httpsServers = make(map[string]*http.Server)
	l.http3Servers = make(map[string]*http3.Server)
}

func (l *lProxy) Reload(cfg model.Cfg) {
	l.state.Set(cfg)
	l.reloadHttpServer(cfg.Http)
	l.reloadHttpsServer(cfg.Https)
	l.reloadHttp3Server(cfg.Http3)
}

func (l *lProxy) reloadHttpServer(cfg *model.HttpCfg) {
	listen := bset.New(cfg.Listen...)
	for k, server := range l.httpServers {
		if listen.Has(k) {
			listen.Remove(k)
			continue
		}
		server.Close()
		delete(l.httpServers, k)
	}

	listen.Range(func(k string) bool {
		server := &http.Server{
			Addr:     k,
			Handler:  l.httpHandler,
			ErrorLog: log.New(io.Discard, "", log.LstdFlags),
		}
		go server.ListenAndServe()
		l.httpServers[k] = server
		return true
	})
}

func (l *lProxy) reloadHttpsServer(cfg *model.HttpsCfg) {
	listen := bset.New(cfg.Listen...)
	for k, server := range l.httpsServers {
		if listen.Has(k) {
			listen.Remove(k)
			continue
		}
		server.Close()
		delete(l.httpsServers, k)
	}

	listen.Range(func(k string) bool {
		server := &http.Server{
			Addr:      k,
			Handler:   l.httpsHandler,
			ErrorLog:  log.New(io.Discard, "", log.LstdFlags),
			TLSConfig: &tls.Config{GetCertificate: l.getHttpsCertificate},
		}
		go server.ListenAndServeTLS("", "")
		l.httpsServers[k] = server
		return true
	})
}
func (l *lProxy) reloadHttp3Server(cfg *model.Http3Cfg) {
	listen := bset.New(cfg.Listen...)
	for k, server := range l.http3Servers {
		if listen.Has(k) {
			listen.Remove(k)
			continue
		}
		server.Close()
		delete(l.http3Servers, k)
	}

	listen.Range(func(k string) bool {
		server := &http3.Server{
			Addr:            k,
			Handler:         l.http3Handler,
			TLSConfig:       &tls.Config{GetCertificate: l.getHttp3Certificate},
			EnableDatagrams: true,
			QuicConfig: &quic.Config{
				EnableDatagrams: true,
				Allow0RTT:       true,
			},
		}
		go server.ListenAndServe()
		l.http3Servers[k] = server
		return true
	})
}

func (l *lProxy) errorHandler(resp http.ResponseWriter, req *http.Request, err error) {
	if err == ErrVhostNotFound {
		resp.WriteHeader(http.StatusForbidden)
		resp.Write(assets.HtmlContentForbidden)
		return
	}
	resp.WriteHeader(http.StatusBadGateway)
}

func (l *lProxy) newGetCertificateFunc(lock *sync.RWMutex, certs map[string]*tls.Certificate) GetCertificateFunc {
	return func(sni *tls.ClientHelloInfo) (*tls.Certificate, error) {
		lock.RLock()
		cert, ok := certs[sni.ServerName]
		lock.RUnlock()
		if !ok {
			return nil, ErrCertNotFound
		}
		return cert, nil
	}
}

func (l *lProxy) timerUpdateOCSP(lock *sync.RWMutex, certs map[string]*tls.Certificate) {
	httpClient := http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   3 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
		Timeout: 3 * time.Second,
	}

	type OCSPRequest struct {
		leaf   *x509.Certificate
		issuer *x509.Certificate
	}

	timer := time.NewTicker(time.Minute)
	for range timer.C {

		requests := make(map[string]OCSPRequest)

		func() {
			lock.RLock()
			defer lock.RUnlock()

			for key, cert := range certs {
				if len(cert.Certificate) < 2 {
					continue
				}

				if len(cert.Leaf.OCSPServer) == 0 {
					continue
				}

				issuer, _ := x509.ParseCertificate(cert.Certificate[1])
				if cert.OCSPStaple == nil {
					requests[key] = OCSPRequest{
						leaf:   cert.Leaf,
						issuer: issuer,
					}
					continue
				}

				resp, err := ocsp.ParseResponse(cert.OCSPStaple, issuer)
				if err != nil || resp.Status != ocsp.Good {
					requests[key] = OCSPRequest{
						leaf:   cert.Leaf,
						issuer: issuer,
					}
					continue
				}

				if time.Until(resp.NextUpdate) < 10*time.Minute {
					requests[key] = OCSPRequest{
						leaf:   cert.Leaf,
						issuer: issuer,
					}
				}
			}
		}()

		if len(requests) == 0 {
			continue
		}
		responses := make(map[string]*ocsp.Response)

		func() {
			eg := &errgroup.Group{}
			eg.SetLimit(10)
			responsesLock := new(sync.Mutex)
			for key, request := range requests {
				eg.Go(func() error {
					der, err := ocsp.CreateRequest(request.leaf, request.issuer, nil)
					if err != nil {
						return err
					}

					uri := request.leaf.OCSPServer[0] + "/" + base64.StdEncoding.EncodeToString(der)
					httpRequest, err := http.NewRequest(http.MethodGet, uri, nil)
					if err != nil {
						return err
					}
					httpRequest.Header.Add("Content-Language", "application/ocsp-request")
					httpRequest.Header.Add("Accept", "application/ocsp-response")
					httpResponse, err := httpClient.Do(httpRequest)
					if err != nil {
						return err
					}
					defer httpResponse.Body.Close()
					der, err = io.ReadAll(httpResponse.Body)
					if err != nil {
						return err
					}

					response, err := ocsp.ParseResponse(der, request.issuer)
					if err != nil {
						return err
					}

					if response.Status == ocsp.Good {
						responsesLock.Lock()
						responses[key] = response
						responsesLock.Unlock()
					}
					return nil
				})
			}

			eg.Wait()
		}()

		if len(responses) == 0 {
			continue
		}

		func() {
			lock.Lock()
			defer lock.Unlock()

			for key, response := range responses {
				cert, ok := certs[key]
				if !ok {
					continue
				}
				cert.OCSPStaple = response.Raw
			}
		}()
	}
}

func (l *lProxy) newReverseProxy(lock *sync.RWMutex, mappings map[string][]*Mapping) http.Handler {
	director := func(req *http.Request) (*http.Response, http.Header, error) {
		lock.RLock()
		mapping, ok := mappings[l.hostname(req)]
		lock.RUnlock()
		if !ok {
			return nil, nil, ErrVhostNotFound
		}

		var t *Mapping
		for _, m := range mapping {
			if strings.HasPrefix(req.URL.Path, m.Path) {
				t = m
				break
			}
		}
		if t == nil {
			return nil, nil, ErrVhostNotFound
		}

		req.URL.Scheme = t.Target.Scheme
		if t.Target.User != nil {
			user := *t.Target.User
			req.URL.User = &user
		}
		req.URL.Host = t.Target.Host

		if t.Redirect {
			header := make(http.Header)
			header.Set("Location", req.URL.String())
			resp := &http.Response{
				Status:        "Moved Permanently",
				StatusCode:    http.StatusMovedPermanently,
				Proto:         req.Proto,
				ProtoMajor:    req.ProtoMajor,
				ProtoMinor:    req.ProtoMinor,
				Header:        header,
				Body:          io.NopCloser(bytes.NewReader(nil)),
				ContentLength: 0,
				Request:       req,
			}
			return resp, t.AddHeader, nil
		}

		// 规则与nginx保持一致
		// 若target是带path信息的
		// 就把原来的Path去掉匹配的前缀，再拼接到Target的Path后面
		if t.Target.Path != "" {
			req.URL.Path = t.Target.Path + req.URL.Path[len(t.Path):]
		}

		if t.ProxyHeader {
			req.Header.Set("X-Forwarded-Proto", l.scheme(req))
		} else {
			req.Host = req.URL.Host
			req.Header.Del("X-Forwarded-For")
		}
		return nil, t.AddHeader, nil
	}

	return &httputil.ReverseProxy{
		Director: func(*http.Request) {},
		Transport: &utils.ReverseProxyTransport{
			Director:       director,
			HttpTransport:  HttpTransport,
			Http3Transport: Http3Transport,
		},
		ErrorLog:     log.New(io.Discard, "", log.LstdFlags),
		ErrorHandler: l.errorHandler,
	}
}

func (l *lProxy) hostname(req *http.Request) string {
	host := req.Host
	end := -1
	for i := len(host) - 1; i >= 0; i-- {
		c := host[i]
		if c == ':' {
			end = i
			break
		}
	}
	if end != -1 {
		host = host[:end]
	}
	return host
}

func (l *lProxy) scheme(req *http.Request) string {
	scheme := strings.ToLower(req.Header.Get("X-Forwarded-Proto"))
	switch scheme {
	case "http", "https":
		return scheme
	}
	if req.TLS == nil {
		return "http"
	}
	return "https"
}
