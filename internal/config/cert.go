package config

import "errors"

type Cert struct {
	Name    string `yaml:"name" json:"name"`
	Content string `yaml:"content" json:"content"`
}

func (c *Cert) CheckValid() error {
	if c.Name == "" {
		return errors.New("cert name missing")
	}
	if c.Content == "" {
		return errors.New("cert content missing")
	}
	return nil
}
