package app

import (
	"errors"

	"github.com/xbugio/go-vhostd/internal/config"
)

func (app *App) DaoGetQuicVhosts() ([]*config.QuicVhost, error) {
	cfg, err := app.DaoGetCfg()
	if err != nil {
		return nil, err
	}

	list := make([]*config.QuicVhost, 0)
	if cfg.Quic != nil && cfg.Quic.Vhost != nil {
		list = cfg.Quic.Vhost
	}
	return list, nil
}

func (app *App) DaoSetQuicVhosts(list []*config.QuicVhost) error {
	cfg, err := app.DaoGetCfg()
	if err != nil {
		return err
	}
	if cfg.Quic == nil {
		cfg.Quic = &config.Quic{}
	}
	cfg.Quic.Vhost = list
	return app.DaoSetCfg(cfg)
}

func (app *App) DaoGetQuicVhost(domain string) (*config.QuicVhost, error) {
	list, err := app.DaoGetQuicVhosts()
	if err != nil {
		return nil, err
	}
	for _, v := range list {
		if v.Domain == domain {
			return v, nil
		}
	}
	return nil, errors.New("quic vhost not found")
}

func (app *App) DaoSetQuicVhost(v *config.QuicVhost) error {
	cfg, err := app.DaoGetCfg()
	if err != nil {
		return err
	}
	if cfg.Quic == nil {
		cfg.Quic = &config.Quic{}
	}
	x := -1
	for i, t := range cfg.Quic.Vhost {
		if t.Domain == v.Domain {
			x = i
			cfg.Quic.Vhost[i] = v
			break
		}
	}
	if x == -1 {
		return errors.New("quic vhost not found")
	}
	return app.DaoSetCfg(cfg)
}

func (app *App) DaoAddQuicVhost(v *config.QuicVhost) error {
	cfg, err := app.DaoGetCfg()
	if err != nil {
		return err
	}
	if cfg.Quic == nil {
		cfg.Quic = &config.Quic{}
	}
	cfg.Quic.Vhost = append(cfg.Quic.Vhost, v)
	return app.DaoSetCfg(cfg)
}

func (app *App) DaoRemoveQuicVhost(domain string) error {
	cfg, err := app.DaoGetCfg()
	if err != nil {
		return err
	}
	if cfg.Quic == nil {
		cfg.Quic = &config.Quic{}
	}
	x := -1
	for i, t := range cfg.Quic.Vhost {
		if t.Domain == domain {
			x = i
			break
		}
	}
	if x == -1 {
		return errors.New("quic vhost not found")
	}
	cfg.Quic.Vhost = append(cfg.Quic.Vhost[:x], cfg.Quic.Vhost[x+1:]...)
	return app.DaoSetCfg(cfg)
}
