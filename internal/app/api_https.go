package app

import (
	"github.com/gin-gonic/gin"
	"github.com/xbugio/go-vhostd/internal/config"
)

func (app *App) ApiGetHttpsConfig(c *gin.Context) {
	app.configLock.RLock()
	defer app.configLock.RUnlock()
	cfg, err := app.DaoGetCfg()
	if err != nil {
		c.Error(err)
		return
	}
	c.Set("data", cfg.Https)
}

func (app *App) ApiSetHttpsConfig(c *gin.Context) {
	httpsCfg := &config.Https{}
	if err := c.BindJSON(httpsCfg); err != nil {
		c.Error(err)
		return
	}

	app.configLock.Lock()
	defer app.configLock.Unlock()

	cfg, err := app.DaoGetCfg()
	if err != nil {
		c.Error(err)
		return
	}
	cfg.Https = httpsCfg
	if err := app.DaoSetCfg(cfg); err != nil {
		c.Error(err)
		return
	}
}

func (app *App) ApiListHttpsVhost(c *gin.Context) {
	app.configLock.RLock()
	defer app.configLock.RUnlock()
	list, err := app.DaoGetHttpsVhosts()
	if err != nil {
		c.Error(err)
		return
	}
	c.Set("data", list)
}

func (app *App) ApiGetHttpsVhost(c *gin.Context) {
	domain := c.Param("domain")

	app.configLock.RLock()
	defer app.configLock.RUnlock()

	data, err := app.DaoGetHttpsVhost(domain)
	if err != nil {
		c.Error(err)
		return
	}
	c.Set("data", data)
}

func (app *App) ApiAddHttpsVhost(c *gin.Context) {
	data := &config.HttpsVhost{}
	if err := c.BindJSON(data); err != nil {
		c.Error(err)
		return
	}

	app.configLock.Lock()
	defer app.configLock.Unlock()

	if err := app.DaoAddHttpsVhost(data); err != nil {
		c.Error(err)
		return
	}
}

func (app *App) ApiDelHttpsVhost(c *gin.Context) {
	domain := c.Param("domain")

	app.configLock.Lock()
	defer app.configLock.Unlock()

	if err := app.DaoRemoveHttpsVhost(domain); err != nil {
		c.Error(err)
		return
	}
}

func (app *App) ApiModHttpsVhost(c *gin.Context) {
	data := &config.HttpsVhost{}
	if err := c.BindJSON(data); err != nil {
		c.Error(err)
		return
	}

	app.configLock.Lock()
	defer app.configLock.Unlock()

	if err := app.DaoSetHttpsVhost(data); err != nil {
		c.Error(err)
		return
	}
}
