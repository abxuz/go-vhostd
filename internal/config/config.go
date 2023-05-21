package config

import (
	"bytes"
	"fmt"
	"io"
	"os"

	yaml "gopkg.in/yaml.v3"
)

type Http struct {
	Listen string       `yaml:"listen" json:"listen"`
	Vhost  []*HttpVhost `yaml:"vhost,omitempty" json:"vhost,omitempty"`
}

type Https struct {
	Listen string        `yaml:"listen" json:"listen"`
	Vhost  []*HttpsVhost `yaml:"vhost,omitempty" json:"vhost,omitempty"`
}

type Quic struct {
	Listen string       `yaml:"listen" json:"listen"`
	Vhost  []*QuicVhost `yaml:"vhost,omitempty" json:"vhost,omitempty"`
}

type Config struct {
	Api   *Api    `yaml:"api,omitempty" json:"api,omitempty"`
	Http  *Http   `yaml:"http,omitempty" json:"http,omitempty"`
	Https *Https  `yaml:"https,omitempty" json:"https,omitempty"`
	Quic  *Quic   `yaml:"quic,omitempty" json:"quic,omitempty"`
	Cert  []*Cert `yaml:"cert,omitempty" json:"cert,omitempty"`
}

func LoadConfigFromBytes(b []byte) (*Config, error) {
	return LoadConfigFromReader(bytes.NewReader(b))
}

func LoadConfigFromFile(p string) (*Config, error) {
	f, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return LoadConfigFromReader(f)
}

func LoadConfigFromReader(r io.Reader) (*Config, error) {
	cfg := &Config{}
	err := yaml.NewDecoder(r).Decode(cfg)
	if err != nil {
		return nil, err
	}
	if err := cfg.CheckValid(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *Config) Encode() ([]byte, error) {
	return yaml.Marshal(c)
}

func (c *Config) EncodeTo(w io.Writer) error {
	return yaml.NewEncoder(w).Encode(c)
}

func (c *Config) WriteFile(p string) error {
	f, err := os.Create(p)
	if err != nil {
		return err
	}
	defer f.Close()
	return c.EncodeTo(f)
}

func (c *Config) CheckValid() error {

	certnames := make(map[string]struct{})
	if c.Cert != nil {
		for _, v := range c.Cert {
			if err := v.CheckValid(); err != nil {
				return err
			}
			if _, ok := certnames[v.Name]; ok {
				return fmt.Errorf("duplicate cert name: %v", v.Name)
			}
			certnames[v.Name] = struct{}{}
		}
	}

	if c.Http != nil {
		domains := make(map[string]struct{})
		for _, v := range c.Http.Vhost {
			if err := v.CheckValid(); err != nil {
				return err
			}
			if _, ok := domains[v.Domain]; ok {
				return fmt.Errorf("duplicate domain in http vhost: %v", v.Domain)
			}
			domains[v.Domain] = struct{}{}
		}
	}

	if c.Https != nil {
		domains := make(map[string]struct{})
		for _, v := range c.Https.Vhost {
			if err := v.CheckValid(); err != nil {
				return err
			}
			if _, ok := certnames[v.Cert]; !ok {
				return fmt.Errorf("cert %v not found in https vhost: %v", v.Cert, v.Name)
			}
			if _, ok := domains[v.Domain]; ok {
				return fmt.Errorf("duplicate domain in https vhost: %v", v.Domain)
			}
			domains[v.Domain] = struct{}{}
		}
	}

	if c.Quic != nil {
		domains := make(map[string]struct{})
		for _, v := range c.Quic.Vhost {
			if err := v.CheckValid(); err != nil {
				return err
			}
			if _, ok := certnames[v.Cert]; !ok {
				return fmt.Errorf("cert %v not found in quic vhost: %v", v.Cert, v.Name)
			}
			if _, ok := domains[v.Domain]; ok {
				return fmt.Errorf("duplicate domain in quic vhost: %v", v.Domain)
			}
			domains[v.Domain] = struct{}{}
		}
	}

	return nil
}
