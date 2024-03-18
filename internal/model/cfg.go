package model

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/abxuz/b-tools/bmap"
	"github.com/abxuz/b-tools/bslice"
	"github.com/abxuz/go-vhostd/utils"
)

type Cfg struct {
	Api   *ApiCfg    `yaml:"api,omitempty" json:"api,omitempty"`
	Http  *HttpCfg   `yaml:"http,omitempty" json:"http,omitempty"`
	Https *HttpsCfg  `yaml:"https,omitempty" json:"https,omitempty"`
	Http3 *Http3Cfg  `yaml:"http3,omitempty" json:"http3,omitempty"`
	Cert  []*CertCfg `yaml:"cert,omitempty" json:"cert,omitempty"`
}

func (c *Cfg) CheckValid() error {
	if c.Http != nil {
		if err := c.Http.CheckValid(); err != nil {
			return err
		}
	}

	if c.Https != nil {
		if err := c.Https.CheckValid(); err != nil {
			return err
		}
	}

	if !bslice.Unique(
		append(c.Http.Listen, c.Https.Listen...),
		func(s string) string { return s },
	) {
		return errors.New("duplicate listen address in http with https config")
	}

	if c.Http3 != nil {
		if err := c.Http3.CheckValid(); err != nil {
			return err
		}
	}

	for _, cert := range c.Cert {
		if err := cert.CheckValid(); err != nil {
			return err
		}
	}

	if !bslice.Unique(c.Cert, func(cert *CertCfg) string { return cert.Name }) {
		return errors.New("duplicate cert name in config")
	}

	certs := bmap.NewMapFromSlice(c.Cert, func(cert *CertCfg) string { return cert.Name })
	if c.Https != nil {
		for _, vhost := range c.Https.Vhost {
			if _, ok := certs[vhost.Cert]; !ok {
				return fmt.Errorf("cert %v not found", vhost.Cert)
			}
		}
	}

	if c.Http3 != nil {
		for _, vhost := range c.Http3.Vhost {
			if _, ok := certs[vhost.Cert]; !ok {
				return fmt.Errorf("cert %v not found", vhost.Cert)
			}
		}
	}

	return nil
}

type ApiCfg struct {
	Listen []string `yaml:"listen" json:"listen"`
	Auth   *AuthCfg `yaml:"auth,omitempty" json:"auth,omitempty"`
}

type AuthCfg struct {
	Username string `yaml:"username" json:"username"`
	Password string `yaml:"password" json:"password"`
}

type HttpCfg struct {
	Listen []string        `yaml:"listen" json:"listen"`
	Vhost  []*HttpVhostCfg `yaml:"vhost,omitempty" json:"vhost,omitempty"`
}

func (c *HttpCfg) CheckValid() error {
	for _, h := range c.Vhost {
		if err := h.CheckValid(); err != nil {
			return err
		}
	}

	if bslice.Unique(c.Vhost, func(h *HttpVhostCfg) string { return h.Domain }) {
		return errors.New("duplicate domain found in vhost config")
	}

	return nil
}

type HttpsCfg struct {
	Listen []string         `yaml:"listen" json:"listen"`
	Vhost  []*HttpsVhostCfg `yaml:"vhost,omitempty" json:"vhost,omitempty"`
}

func (c *HttpsCfg) CheckValid() error {
	for _, h := range c.Vhost {
		if err := h.CheckValid(); err != nil {
			return err
		}
	}

	if bslice.Unique(c.Vhost, func(h *HttpsVhostCfg) string { return h.Domain }) {
		return errors.New("duplicate domain found in vhost config")
	}

	return nil
}

type Http3Cfg struct {
	Listen []string         `yaml:"listen" json:"listen"`
	Vhost  []*Http3VhostCfg `yaml:"vhost,omitempty" json:"vhost,omitempty"`
}

func (c *Http3Cfg) CheckValid() error {
	for _, h := range c.Vhost {
		if err := h.CheckValid(); err != nil {
			return err
		}
	}

	if bslice.Unique(c.Vhost, func(h *Http3VhostCfg) string { return h.Domain }) {
		return errors.New("duplicate domain found in vhost config")
	}

	return nil
}

type HttpVhostCfg struct {
	VhostCfg `yaml:",inline"`
}

type HttpsVhostCfg struct {
	VhostCfg `yaml:",inline"`
	Cert     string `yaml:"cert" json:"cert"`
}

func (c *HttpsVhostCfg) CheckValid() error {
	if err := c.VhostCfg.CheckValid(); err != nil {
		return err
	}

	if utils.ExistEmptyString(true, c.Cert) {
		return errors.New("cert required for vhost config")
	}

	return nil
}

type Http3VhostCfg struct {
	VhostCfg `yaml:",inline"`
	Cert     string `yaml:"cert" json:"cert"`
}

func (c *Http3VhostCfg) CheckValid() error {
	if err := c.VhostCfg.CheckValid(); err != nil {
		return err
	}

	if utils.ExistEmptyString(true, c.Cert) {
		return errors.New("cert required for vhost config")
	}

	return nil
}

type VhostCfg struct {
	Name    string        `yaml:"name" json:"name"`
	Domain  string        `yaml:"domain" json:"domain"`
	Mapping []*MappingCfg `yaml:"mapping" json:"mapping"`
}

func (c *VhostCfg) CheckValid() error {
	if utils.ExistEmptyString(true, c.Domain) {
		return errors.New("domain required for vhost config")
	}

	if len(c.Mapping) == 0 {
		return errors.New("mapping required for vhost config")
	}

	for _, m := range c.Mapping {
		if err := m.CheckValid(); err != nil {
			return err
		}
	}

	if !bslice.Unique(c.Mapping, func(c *MappingCfg) string { return c.Path }) {
		return errors.New("duplicate mapping path in vhost config")
	}

	return nil
}

type MappingCfg struct {
	Path        string   `yaml:"path" json:"path"`
	Target      string   `yaml:"target" json:"target"`
	AddHeader   []string `yaml:"add_header" json:"add_header"`
	ProxyHeader bool     `yaml:"proxy_header" json:"proxy_header"`
	Redirect    bool     `yaml:"redirect" json:"redirect"`
}

func (c *MappingCfg) CheckValid() error {
	if utils.ExistEmptyString(true, c.Path) {
		return errors.New("path required for vhost mapping config")
	}

	_, err := c.GetTarget()
	if err != nil {
		return err
	}

	_, err = c.GetAddHeader()
	return err
}

func (c *MappingCfg) GetTarget() (*url.URL, error) {
	u, err := url.ParseRequestURI(c.Target)
	if err != nil {
		return nil, err
	}
	if utils.ExistEmptyString(false, u.Scheme, u.Host) {
		return nil, errors.New("malform target, missing scheme or host")
	}
	return u, nil
}

func (c *MappingCfg) GetAddHeader() (http.Header, error) {
	header := make(http.Header)
	for _, h := range c.AddHeader {
		items := strings.SplitN(h, ":", 2)
		if len(items) != 2 {
			return nil, errors.New("malform add_header")
		}
		k := strings.TrimSpace(items[0])
		v := strings.TrimSpace(items[1])
		header.Add(k, v)
	}
	return header, nil
}

type CertCfg struct {
	Name    string `yaml:"name" json:"name"`
	Content string `yaml:"content" json:"content"`
}

func (c *CertCfg) Certificate() (*tls.Certificate, error) {
	cert, err := utils.ParseCert([]byte(c.Content))
	if err != nil {
		return nil, err
	}
	if len(cert.Certificate) == 0 {
		return nil, errors.New("no certificate found")
	}
	return cert, nil
}

func (c *CertCfg) CertInfo() (*CertInfo, error) {
	cert, err := c.Certificate()
	if err != nil {
		return nil, err
	}

	i, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		return nil, err
	}

	return &CertInfo{
		Domain:     i.DNSNames,
		Issuer:     i.Issuer.String(),
		ValidStart: i.NotBefore.Local().Format(time.DateTime),
		ValidStop:  i.NotAfter.Local().Format(time.DateTime),
	}, nil
}

func (c *CertCfg) CheckValid() error {
	if utils.ExistEmptyString(true, c.Name, c.Content) {
		return errors.New("name or content required for cert config")
	}
	_, err := c.Certificate()
	return err
}
