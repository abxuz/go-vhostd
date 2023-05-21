package app

import "github.com/xbugio/go-vhostd/internal/config"

func (app *App) DaoGetCfg() (*config.Config, error) {
	cfg, err := config.LoadConfigFromFile(app.config)
	if err != nil {
		return nil, err
	}
	if err := cfg.CheckValid(); err != nil {
		return nil, err
	}
	if _, err := GenState(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (app *App) DaoSetCfg(cfg *config.Config) error {
	if err := cfg.CheckValid(); err != nil {
		return err
	}
	if _, err := GenState(cfg); err != nil {
		return err
	}
	return cfg.WriteFile(app.config)
}
