package app

import (
	"bytes"
	"crypto/tls"
	"errors"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"

	"github.com/quic-go/quic-go/http3"
)

var (
	DefaultTransport = &http.Transport{
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
	LogDiscard  = log.New(io.Discard, "", log.LstdFlags)
	NopDirector = func(*http.Request) {}

	ErrVhostNotFound = errors.New("vhost not found")
	ErrCertNotFound  = errors.New("cert not found")
)

type ReverseProxyTransport struct {
	http.RoundTripper
	Director func(*http.Request) error
}

func (t *ReverseProxyTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if err := t.Director(req); err != nil {
		return nil, err
	}
	if req.Response != nil {
		return req.Response, nil
	}
	return t.RoundTripper.RoundTrip(req)
}

func (app *App) newHttpServer(state *HttpState) *http.Server {
	handler := &httputil.ReverseProxy{
		Director: NopDirector,
		Transport: &ReverseProxyTransport{
			RoundTripper: DefaultTransport,
			Director:     app.httpDirector,
		},
		ErrorLog:     LogDiscard,
		ErrorHandler: app.handleProxyError,
	}
	return &http.Server{
		Addr:     state.Listen,
		Handler:  handler,
		ErrorLog: LogDiscard,
	}
}

func (app *App) newHttpsServer(state *HttpsState) *http.Server {
	handler := &httputil.ReverseProxy{
		Director: NopDirector,
		Transport: &ReverseProxyTransport{
			RoundTripper: DefaultTransport,
			Director:     app.httpsDirector,
		},
		ErrorLog:     LogDiscard,
		ErrorHandler: app.handleProxyError,
	}
	return &http.Server{
		Addr:     state.Listen,
		Handler:  handler,
		ErrorLog: LogDiscard,
		TLSConfig: &tls.Config{
			GetCertificate: app.getHttpsCertificate,
		},
	}
}

func (app *App) newQuicServer(state *QuicState) *http3.Server {
	handler := &httputil.ReverseProxy{
		Director: NopDirector,
		Transport: &ReverseProxyTransport{
			RoundTripper: DefaultTransport,
			Director:     app.quicDirector,
		},
		ErrorLog:     LogDiscard,
		ErrorHandler: app.handleProxyError,
	}
	return &http3.Server{
		Addr:    state.Listen,
		Handler: handler,
		TLSConfig: &tls.Config{
			GetCertificate: app.getQuicCertificate,
		},
	}
}

func (app *App) httpDirector(req *http.Request) error {
	app.httpStateLock.RLock()
	defer app.httpStateLock.RUnlock()

	if app.httpState == nil {
		return ErrVhostNotFound
	}

	domain := hostname(req)
	vhost, exists := app.httpState.Vhost[domain]
	if !exists {
		return ErrVhostNotFound
	}

	return app.vhostDirector(req, vhost.Mapping, "http")
}

func (app *App) httpsDirector(req *http.Request) error {
	app.httpStateLock.RLock()
	defer app.httpStateLock.RUnlock()

	if app.httpsState == nil {
		return ErrVhostNotFound
	}

	domain := hostname(req)
	vhost, exists := app.httpsState.Vhost[domain]
	if !exists {
		return ErrVhostNotFound
	}

	return app.vhostDirector(req, vhost.Mapping, "https")
}

func (app *App) quicDirector(req *http.Request) error {
	app.quicStateLock.RLock()
	defer app.quicStateLock.RUnlock()

	if app.quicState == nil {
		return ErrVhostNotFound
	}

	domain := hostname(req)
	vhost, exists := app.quicState.Vhost[domain]
	if !exists {
		return ErrVhostNotFound
	}

	return app.vhostDirector(req, vhost.Mapping, "https")
}

func (app *App) vhostDirector(req *http.Request, mapping []*Mapping, proto string) error {
	var t *Mapping
	for _, m := range mapping {
		if strings.HasPrefix(req.URL.Path, m.Path) {
			t = m
			break
		}
	}
	if t == nil {
		return ErrVhostNotFound
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
		req.Response = &http.Response{
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
		return nil
	}

	// 规则与nginx保持一致
	// 若target是带path信息的
	// 就把原来的Path去掉匹配的前缀，再拼接到Target的Path后面
	if t.Target.Path != "" {
		req.URL.Path = t.Target.Path + req.URL.Path[len(t.Path):]
	}

	if t.ProxyHeader {
		req.Header.Set("X-Forwarded-Proto", proto)
	} else {
		req.Host = req.URL.Host
		req.Header.Del("X-Forwarded-For")
	}

	for _, h := range t.AddHeader {
		req.Header.Add(h.Key, h.Value)
	}

	return nil
}

func (app *App) getHttpsCertificate(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
	var certname string

	app.httpsStateLock.RLock()
	if app.httpsState != nil {
		vhost, ok := app.httpsState.Vhost[hello.ServerName]
		if ok {
			certname = vhost.Cert
		}
	}
	app.httpsStateLock.RUnlock()

	return app.getCertificate(certname)
}

func (app *App) getQuicCertificate(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
	var certname string

	app.quicStateLock.RLock()
	if app.quicState != nil {
		vhost, ok := app.quicState.Vhost[hello.ServerName]
		if ok {
			certname = vhost.Cert
		}
	}
	app.quicStateLock.RUnlock()

	return app.getCertificate(certname)
}

func (app *App) getCertificate(certname string) (*tls.Certificate, error) {
	app.certStateLock.RLock()
	defer app.certStateLock.RUnlock()

	if app.certState == nil || app.certState.Default == nil {
		return nil, ErrCertNotFound
	}

	if len(app.certState.Cert) == 0 {
		return app.certState.Default.Certificate, nil
	}

	cert, ok := app.certState.Cert[certname]
	if ok {
		return cert.Certificate, nil
	}

	return app.certState.Default.Certificate, nil
}

func (app *App) handleProxyError(rw http.ResponseWriter, req *http.Request, err error) {
	if err == ErrVhostNotFound {
		app.ServeDefaultVhost(rw, req)
		return
	}
	rw.WriteHeader(http.StatusBadGateway)
}

func (app *App) ServeDefaultVhost(rw http.ResponseWriter, req *http.Request) {
	rw.WriteHeader(http.StatusForbidden)
	rw.Write(HtmlContentForbidden)
}

func hostname(req *http.Request) string {
	host := req.Host
	end := strings.Index(req.Host, ":")
	if end != -1 {
		host = host[:end]
	}
	return host
}
