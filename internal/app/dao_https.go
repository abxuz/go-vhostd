package app

import (
	"errors"

	"github.com/xbugio/go-vhostd/internal/config"
)

func (app *App) DaoGetHttpsVhosts() ([]*config.HttpsVhost, error) {
	cfg, err := app.DaoGetCfg()
	if err != nil {
		return nil, err
	}

	list := make([]*config.HttpsVhost, 0)
	if cfg.Https != nil && cfg.Https.Vhost != nil {
		list = cfg.Https.Vhost
	}
	return list, nil
}

func (app *App) DaoSetHttpsVhosts(list []*config.HttpsVhost) error {
	cfg, err := app.DaoGetCfg()
	if err != nil {
		return err
	}
	if cfg.Https == nil {
		cfg.Https = &config.Https{}
	}
	cfg.Https.Vhost = list
	return app.DaoSetCfg(cfg)
}

func (app *App) DaoGetHttpsVhost(domain string) (*config.HttpsVhost, error) {
	list, err := app.DaoGetHttpsVhosts()
	if err != nil {
		return nil, err
	}
	for _, v := range list {
		if v.Domain == domain {
			return v, nil
		}
	}
	return nil, errors.New("https vhost not found")
}

func (app *App) DaoSetHttpsVhost(v *config.HttpsVhost) error {
	cfg, err := app.DaoGetCfg()
	if err != nil {
		return err
	}
	if cfg.Https == nil {
		cfg.Https = &config.Https{}
	}
	x := -1
	for i, t := range cfg.Https.Vhost {
		if t.Domain == v.Domain {
			x = i
			cfg.Https.Vhost[i] = v
			break
		}
	}
	if x == -1 {
		return errors.New("https vhost not found")
	}
	return app.DaoSetCfg(cfg)
}

func (app *App) DaoAddHttpsVhost(v *config.HttpsVhost) error {
	cfg, err := app.DaoGetCfg()
	if err != nil {
		return err
	}
	if cfg.Https == nil {
		cfg.Https = &config.Https{}
	}
	cfg.Https.Vhost = append(cfg.Https.Vhost, v)
	return app.DaoSetCfg(cfg)
}

func (app *App) DaoRemoveHttpsVhost(domain string) error {
	cfg, err := app.DaoGetCfg()
	if err != nil {
		return err
	}
	if cfg.Https == nil {
		cfg.Https = &config.Https{}
	}
	x := -1
	for i, t := range cfg.Https.Vhost {
		if t.Domain == domain {
			x = i
			break
		}
	}
	if x == -1 {
		return errors.New("https vhost not found")
	}
	cfg.Https.Vhost = append(cfg.Https.Vhost[:x], cfg.Https.Vhost[x+1:]...)
	return app.DaoSetCfg(cfg)
}
