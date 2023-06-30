package app

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net/http"
	"net/url"

	"github.com/xbugio/go-vhostd/internal/config"
	"github.com/xbugio/go-vhostd/internal/fs"
)

type VhostListen struct {
	Http  string `json:"http"`
	Https string `json:"https"`
	Quic  string `json:"quic"`
}

type CertInfo struct {
	Domain     []string `json:"domain"`
	Issuer     string   `json:"issuer"`
	ValidStart string   `json:"valid_start"`
	ValidStop  string   `json:"valid_stop"`
}

type ApiState struct {
	Listen            string
	Auth              *config.BasicAuth
	PublicFileHandler http.Handler
}

type Mapping struct {
	Path        string
	Target      *url.URL
	ProxyHeader bool
	Redirect    bool
}

type Vhost struct {
	Name    string
	Domain  string
	Mapping []*Mapping
}

type HttpVhost struct {
	Vhost
}

type HttpsVhost struct {
	Vhost
	Cert string
}

type QuicVhost struct {
	Vhost
	Cert string
}

type HttpState struct {
	Listen string
	Vhost  map[string]*HttpVhost
}

type HttpsState struct {
	Listen string
	Vhost  map[string]*HttpsVhost
}

type QuicState struct {
	Listen string
	Vhost  map[string]*QuicVhost
}

type Cert struct {
	Name        string
	Certificate *tls.Certificate
}

type CertState struct {
	Default *Cert
	Cert    map[string]*Cert
}

type State struct {
	Api   *ApiState
	Http  *HttpState
	Https *HttpsState
	Quic  *QuicState
	Cert  *CertState
}

func GenApiState(cfg *config.Api) (*ApiState, error) {
	state := &ApiState{}
	state.Listen = cfg.Listen
	state.Auth = cfg.Auth
	if cfg.Public != nil {
		state.PublicFileHandler = http.FileServer(&fs.NoAutoIndexFileSystem{
			FileSystem: http.Dir(*cfg.Public),
		})
	}
	return state, nil
}

func GenMapping(cfg *config.Mapping) (*Mapping, error) {
	target, err := url.ParseRequestURI(cfg.Target)
	if err != nil {
		return nil, err
	}
	return &Mapping{
		Path:        cfg.Path,
		Target:      target,
		ProxyHeader: cfg.ProxyHeader,
		Redirect:    cfg.Redirect,
	}, nil
}

func GenHttpVhost(cfg *config.HttpVhost) (*HttpVhost, error) {
	vhost := &HttpVhost{}
	vhost.Name = cfg.Name
	vhost.Domain = cfg.Domain
	vhost.Mapping = make([]*Mapping, 0)
	for _, mCfg := range cfg.Mapping {
		m, err := GenMapping(mCfg)
		if err != nil {
			return nil, err
		}
		vhost.Mapping = append(vhost.Mapping, m)
	}
	return vhost, nil
}

func GenHttpsVhost(cfg *config.HttpsVhost) (*HttpsVhost, error) {
	vhost := &HttpsVhost{}
	vhost.Name = cfg.Name
	vhost.Domain = cfg.Domain
	vhost.Mapping = make([]*Mapping, 0)
	for _, mCfg := range cfg.Mapping {
		m, err := GenMapping(mCfg)
		if err != nil {
			return nil, err
		}
		vhost.Mapping = append(vhost.Mapping, m)
	}
	vhost.Cert = cfg.Cert
	return vhost, nil
}

func GenQuicVhost(cfg *config.QuicVhost) (*QuicVhost, error) {
	vhost := &QuicVhost{}
	vhost.Name = cfg.Name
	vhost.Domain = cfg.Domain
	vhost.Mapping = make([]*Mapping, 0)
	for _, mCfg := range cfg.Mapping {
		m, err := GenMapping(mCfg)
		if err != nil {
			return nil, err
		}
		vhost.Mapping = append(vhost.Mapping, m)
	}
	vhost.Cert = cfg.Cert
	return vhost, nil
}

func GenHttpState(cfg *config.Http) (*HttpState, error) {
	h := &HttpState{}
	h.Listen = cfg.Listen
	h.Vhost = make(map[string]*HttpVhost)
	for _, vCfg := range cfg.Vhost {
		vhost, err := GenHttpVhost(vCfg)
		if err != nil {
			return nil, err
		}
		h.Vhost[vhost.Domain] = vhost
	}
	return h, nil
}

func GenHttpsState(cfg *config.Https) (*HttpsState, error) {
	h := &HttpsState{}
	h.Listen = cfg.Listen
	h.Vhost = make(map[string]*HttpsVhost)
	for _, vCfg := range cfg.Vhost {
		vhost, err := GenHttpsVhost(vCfg)
		if err != nil {
			return nil, err
		}
		h.Vhost[vhost.Domain] = vhost
	}
	return h, nil
}

func GenQuicState(cfg *config.Quic) (*QuicState, error) {
	h := &QuicState{}
	h.Listen = cfg.Listen
	h.Vhost = make(map[string]*QuicVhost)
	for _, vCfg := range cfg.Vhost {
		vhost, err := GenQuicVhost(vCfg)
		if err != nil {
			return nil, err
		}
		h.Vhost[vhost.Domain] = vhost
	}
	return h, nil
}

func GenCert(cfg *config.Cert) (*Cert, error) {
	data := []byte(cfg.Content)

	cert := &tls.Certificate{}
	for {
		var block *pem.Block
		block, data = pem.Decode(data)
		if block == nil {
			break
		}

		switch block.Type {
		case "CERTIFICATE":
			cert.Certificate = append(cert.Certificate, block.Bytes)
		case "RSA PRIVATE KEY":
			pk, err := x509.ParsePKCS1PrivateKey(block.Bytes)
			if err != nil {
				return nil, err
			}
			cert.PrivateKey = pk
		}
	}

	if cert.PrivateKey == nil || len(cert.Certificate) == 0 {
		return nil, fmt.Errorf("invalid cert content for name: %v", cfg.Name)
	}

	return &Cert{Name: cfg.Name, Certificate: cert}, nil
}

func GenCertState(certs []*config.Cert) (*CertState, error) {
	state := &CertState{}
	state.Cert = make(map[string]*Cert)
	for _, vCfg := range certs {
		cert, err := GenCert(vCfg)
		if err != nil {
			return nil, err
		}
		if state.Default == nil {
			state.Default = cert
		}
		state.Cert[cert.Name] = cert
	}
	return state, nil
}

func GenState(cfg *config.Config) (*State, error) {
	var err error

	state := &State{}
	if cfg.Api != nil {
		state.Api, err = GenApiState(cfg.Api)
		if err != nil {
			return nil, err
		}
	}

	if cfg.Http != nil {
		state.Http, err = GenHttpState(cfg.Http)
		if err != nil {
			return nil, err
		}
	}

	if cfg.Https != nil {
		state.Https, err = GenHttpsState(cfg.Https)
		if err != nil {
			return nil, err
		}
	}

	if cfg.Quic != nil {
		state.Quic, err = GenQuicState(cfg.Quic)
		if err != nil {
			return nil, err
		}
	}

	if len(cfg.Cert) > 0 {
		state.Cert, err = GenCertState(cfg.Cert)
		if err != nil {
			return nil, err
		}
	}

	return state, nil
}
