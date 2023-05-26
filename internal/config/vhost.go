package config

import (
	"errors"
	"fmt"
	"net/url"
)

type Mapping struct {
	Path        string `yaml:"path" json:"path"`
	Target      string `yaml:"target" json:"target"`
	ProxyHeader bool   `yaml:"proxy_header" json:"proxy_header"`
}

type Vhost struct {
	Name    string     `yaml:"name" json:"name"`
	Domain  string     `yaml:"domain" json:"domain"`
	Mapping []*Mapping `yaml:"mapping" json:"mapping"`
}

type HttpVhost struct {
	Vhost `yaml:",inline"`
}

type HttpsVhost struct {
	Vhost `yaml:",inline"`
	Cert  string `yaml:"cert" json:"cert"`
}

type QuicVhost struct {
	Vhost `yaml:",inline"`
	Cert  string `yaml:"cert" json:"cert"`
}

func (m *Mapping) CheckValid() error {
	if m.Path == "" {
		return errors.New("mapping path is empty")
	}

	u, err := url.Parse(m.Target)
	if err != nil {
		return err
	}
	if u.Scheme == "" || u.Host == "" {
		return fmt.Errorf("mapping target invalid: %v", m.Target)
	}
	return nil
}

func (v *Vhost) checkValid(t string) error {
	if v.Name == "" {
		return fmt.Errorf("%v vhost name missing", t)
	}
	if v.Domain == "" {
		return fmt.Errorf("%v vhost %v domain missing", t, v.Name)
	}
	if len(v.Mapping) == 0 {
		return fmt.Errorf("%v vhost %v mapping missing", t, v.Name)
	}
	paths := make(map[string]struct{})
	for _, m := range v.Mapping {
		if err := m.CheckValid(); err != nil {
			return err
		}
		if _, ok := paths[m.Path]; ok {
			return fmt.Errorf("duplicate path %v for %v vhost %v", m.Path, t, v.Name)
		}
		paths[m.Path] = struct{}{}
	}
	return nil
}

func (v *HttpVhost) CheckValid() error {
	return v.Vhost.checkValid("http")
}

func (v *HttpsVhost) CheckValid() error {
	if err := v.Vhost.checkValid("https"); err != nil {
		return err
	}
	if v.Cert == "" {
		return fmt.Errorf("https vhost %v cert missing", v.Name)
	}
	return nil
}

func (v *QuicVhost) CheckValid() error {
	if err := v.Vhost.checkValid("quic"); err != nil {
		return err
	}
	if v.Cert == "" {
		return fmt.Errorf("quic vhost %v cert missing", v.Name)
	}
	return nil
}
