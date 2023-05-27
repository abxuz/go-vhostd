package app

import (
	"github.com/gin-gonic/gin"
	"github.com/xbugio/go-vhostd/internal/config"
)

func (app *App) ApiGetHttpConfig(c *gin.Context) {
	app.configLock.RLock()
	defer app.configLock.RUnlock()
	cfg, err := app.DaoGetCfg()
	if err != nil {
		c.Error(err)
		return
	}
	c.Set("data", cfg.Http)
}

func (app *App) ApiSetHttpConfig(c *gin.Context) {
	httpCfg := &config.Http{}
	if err := c.ShouldBindJSON(httpCfg); err != nil {
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
	cfg.Http = httpCfg
	if err := app.DaoSetCfg(cfg); err != nil {
		c.Error(err)
		return
	}
}

func (app *App) ApiListHttpVhost(c *gin.Context) {
	app.configLock.RLock()
	defer app.configLock.RUnlock()
	list, err := app.DaoGetHttpVhosts()
	if err != nil {
		c.Error(err)
		return
	}
	c.Set("data", list)
}

func (app *App) ApiGetHttpVhost(c *gin.Context) {
	domain := c.Param("domain")

	app.configLock.RLock()
	defer app.configLock.RUnlock()

	data, err := app.DaoGetHttpVhost(domain)
	if err != nil {
		c.Error(err)
		return
	}
	c.Set("data", data)
}

func (app *App) ApiAddHttpVhost(c *gin.Context) {
	data := &config.HttpVhost{}
	if err := c.ShouldBindJSON(data); err != nil {
		c.Error(err)
		return
	}

	app.configLock.Lock()
	defer app.configLock.Unlock()

	if err := app.DaoAddHttpVhost(data); err != nil {
		c.Error(err)
		return
	}
}

func (app *App) ApiDelHttpVhost(c *gin.Context) {
	domain := c.Param("domain")

	app.configLock.Lock()
	defer app.configLock.Unlock()

	if err := app.DaoRemoveHttpVhost(domain); err != nil {
		c.Error(err)
		return
	}
}

func (app *App) ApiModHttpVhost(c *gin.Context) {
	data := &config.HttpVhost{}
	if err := c.ShouldBindJSON(data); err != nil {
		c.Error(err)
		return
	}

	app.configLock.Lock()
	defer app.configLock.Unlock()

	if err := app.DaoSetHttpVhost(data); err != nil {
		c.Error(err)
		return
	}
}
