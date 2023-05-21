package app

import (
	"crypto/x509"
	"errors"
	"time"

	"github.com/xbugio/go-vhostd/internal/config"
)

func (app *App) DaoGetCerts() ([]*config.Cert, error) {
	cfg, err := app.DaoGetCfg()
	if err != nil {
		return nil, err
	}
	list := make([]*config.Cert, 0)
	if cfg.Cert != nil {
		list = cfg.Cert
	}
	return list, nil
}

func (app *App) DaoSetCerts(list []*config.Cert) error {
	cfg, err := app.DaoGetCfg()
	if err != nil {
		return err
	}
	cfg.Cert = list
	return app.DaoSetCfg(cfg)
}

func (app *App) DaoGetCert(name string) (*config.Cert, error) {
	list, err := app.DaoGetCerts()
	if err != nil {
		return nil, err
	}
	for _, v := range list {
		if v.Name == name {
			return v, nil
		}
	}
	return nil, errors.New("cert not found")
}

func (app *App) DaoSetCert(v *config.Cert) error {
	cfg, err := app.DaoGetCfg()
	if err != nil {
		return err
	}
	x := -1
	for i, t := range cfg.Cert {
		if t.Name == v.Name {
			x = i
			cfg.Cert[i] = v
			break
		}
	}
	if x == -1 {
		return errors.New("cert not found")
	}
	return app.DaoSetCfg(cfg)
}

func (app *App) DaoAddCert(v *config.Cert) error {
	cfg, err := app.DaoGetCfg()
	if err != nil {
		return err
	}
	cfg.Cert = append(cfg.Cert, v)
	return app.DaoSetCfg(cfg)
}

func (app *App) DaoRemoveCert(name string) error {
	cfg, err := app.DaoGetCfg()
	if err != nil {
		return err
	}
	x := -1
	for i, t := range cfg.Cert {
		if t.Name == name {
			x = i
			break
		}
	}
	if x == -1 {
		return errors.New("cert not found")
	}
	cfg.Cert = append(cfg.Cert[:x], cfg.Cert[x+1:]...)
	return app.DaoSetCfg(cfg)
}

func (app *App) DaoGetCertInfo(cfg *config.Cert) (*CertInfo, error) {
	state, err := GenCert(cfg)
	if err != nil {
		return nil, err
	}

	cert, err := x509.ParseCertificate(state.Certificate.Certificate[0])
	if err != nil {
		return nil, err
	}

	info := &CertInfo{}
	info.Domain = cert.DNSNames
	info.Issuer = cert.Issuer.String()
	info.ValidStart = cert.NotBefore.Local().Format(time.DateTime)
	info.ValidStop = cert.NotAfter.Local().Format(time.DateTime)

	return info, nil
}
