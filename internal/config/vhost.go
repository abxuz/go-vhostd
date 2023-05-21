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

type HttpVhost struct {
	Name    string     `yaml:"name" json:"name"`
	Domain  string     `yaml:"domain" json:"domain"`
	Mapping []*Mapping `yaml:"mapping" json:"mapping"`
}

type HttpsVhost struct {
	Name    string     `yaml:"name" json:"name"`
	Domain  string     `yaml:"domain" json:"domain"`
	Mapping []*Mapping `yaml:"mapping" json:"mapping"`
	Cert    string     `yaml:"cert" json:"cert"`
}

type QuicVhost struct {
	Name    string     `yaml:"name" json:"name"`
	Domain  string     `yaml:"domain" json:"domain"`
	Mapping []*Mapping `yaml:"mapping" json:"mapping"`
	Cert    string     `yaml:"cert" json:"cert"`
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

func (v *HttpVhost) CheckValid() error {
	if v.Name == "" {
		return errors.New("http vhost name missing")
	}

	if v.Domain == "" {
		return fmt.Errorf("http vhost %v domain missing", v.Name)
	}

	if len(v.Mapping) == 0 {
		return fmt.Errorf("http vhost %v mapping missing", v.Name)
	}

	paths := make(map[string]struct{})
	for _, m := range v.Mapping {
		if err := m.CheckValid(); err != nil {
			return err
		}
		if _, ok := paths[m.Path]; ok {
			return fmt.Errorf("duplicate path %v for http vhost %v", m.Path, v.Name)
		}
		paths[m.Path] = struct{}{}
	}
	return nil
}

func (v *HttpsVhost) CheckValid() error {
	if v.Name == "" {
		return errors.New("https vhost name missing")
	}

	if v.Domain == "" {
		return fmt.Errorf("https vhost %v domain missing", v.Name)
	}

	if len(v.Mapping) == 0 {
		return fmt.Errorf("https vhost %v mapping missing", v.Name)
	}

	paths := make(map[string]struct{})
	for _, m := range v.Mapping {
		if err := m.CheckValid(); err != nil {
			return err
		}
		if _, ok := paths[m.Path]; ok {
			return fmt.Errorf("duplicate path %v in https vhost %v", m.Path, v.Name)
		}
		paths[m.Path] = struct{}{}
	}

	if v.Cert == "" {
		return fmt.Errorf("https vhost %v cert missing", v.Name)
	}
	return nil
}

func (v *QuicVhost) CheckValid() error {
	if v.Name == "" {
		return errors.New("quic vhost name missing")
	}

	if v.Domain == "" {
		return fmt.Errorf("quic vhost %v domain missing", v.Name)
	}

	if len(v.Mapping) == 0 {
		return fmt.Errorf("quic vhost %v mapping missing", v.Name)
	}

	paths := make(map[string]struct{})
	for _, m := range v.Mapping {
		if err := m.CheckValid(); err != nil {
			return err
		}
		if _, ok := paths[m.Path]; ok {
			return fmt.Errorf("duplicate mapping path %v in quic vhost %v", m.Path, v.Name)
		}
		paths[m.Path] = struct{}{}
	}

	if v.Cert == "" {
		return fmt.Errorf("quic vhost %v cert missing", v.Name)
	}
	return nil
}
