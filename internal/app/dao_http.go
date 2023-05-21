package app

import (
	"errors"

	"github.com/xbugio/go-vhostd/internal/config"
)

func (app *App) DaoGetHttpVhosts() ([]*config.HttpVhost, error) {
	cfg, err := app.DaoGetCfg()
	if err != nil {
		return nil, err
	}

	list := make([]*config.HttpVhost, 0)
	if cfg.Http != nil && cfg.Http.Vhost != nil {
		list = cfg.Http.Vhost
	}
	return list, nil
}

func (app *App) DaoSetHttpVhosts(list []*config.HttpVhost) error {
	cfg, err := app.DaoGetCfg()
	if err != nil {
		return err
	}
	if cfg.Http == nil {
		cfg.Http = &config.Http{}
	}
	cfg.Http.Vhost = list
	return app.DaoSetCfg(cfg)
}

func (app *App) DaoGetHttpVhost(domain string) (*config.HttpVhost, error) {
	list, err := app.DaoGetHttpVhosts()
	if err != nil {
		return nil, err
	}
	for _, v := range list {
		if v.Domain == domain {
			return v, nil
		}
	}
	return nil, errors.New("http vhost not found")
}

func (app *App) DaoSetHttpVhost(v *config.HttpVhost) error {
	cfg, err := app.DaoGetCfg()
	if err != nil {
		return err
	}
	if cfg.Http == nil {
		cfg.Http = &config.Http{}
	}
	x := -1
	for i, t := range cfg.Http.Vhost {
		if t.Domain == v.Domain {
			x = i
			cfg.Http.Vhost[i] = v
			break
		}
	}
	if x == -1 {
		return errors.New("http vhost not found")
	}
	return app.DaoSetCfg(cfg)
}

func (app *App) DaoAddHttpVhost(v *config.HttpVhost) error {
	cfg, err := app.DaoGetCfg()
	if err != nil {
		return err
	}
	if cfg.Http == nil {
		cfg.Http = &config.Http{}
	}
	cfg.Http.Vhost = append(cfg.Http.Vhost, v)
	return app.DaoSetCfg(cfg)
}

func (app *App) DaoRemoveHttpVhost(domain string) error {
	cfg, err := app.DaoGetCfg()
	if err != nil {
		return err
	}
	if cfg.Http == nil {
		cfg.Http = &config.Http{}
	}
	x := -1
	for i, t := range cfg.Http.Vhost {
		if t.Domain == domain {
			x = i
			break
		}
	}
	if x == -1 {
		return errors.New("http vhost not found")
	}
	cfg.Http.Vhost = append(cfg.Http.Vhost[:x], cfg.Http.Vhost[x+1:]...)
	return app.DaoSetCfg(cfg)
}
