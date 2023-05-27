package app

import (
	"github.com/gin-gonic/gin"
	"github.com/xbugio/go-vhostd/internal/config"
)

func (app *App) ApiGetQuicConfig(c *gin.Context) {
	app.configLock.RLock()
	defer app.configLock.RUnlock()
	cfg, err := app.DaoGetCfg()
	if err != nil {
		c.Error(err)
		return
	}
	c.Set("data", cfg.Quic)
}

func (app *App) ApiSetQuicConfig(c *gin.Context) {
	quicCfg := &config.Quic{}
	if err := c.ShouldBindJSON(quicCfg); err != nil {
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
	cfg.Quic = quicCfg
	if err := app.DaoSetCfg(cfg); err != nil {
		c.Error(err)
		return
	}
}

func (app *App) ApiListQuicVhost(c *gin.Context) {
	app.configLock.RLock()
	defer app.configLock.RUnlock()
	list, err := app.DaoGetQuicVhosts()
	if err != nil {
		c.Error(err)
		return
	}
	c.Set("data", list)
}

func (app *App) ApiGetQuicVhost(c *gin.Context) {
	domain := c.Param("domain")

	app.configLock.RLock()
	defer app.configLock.RUnlock()

	data, err := app.DaoGetQuicVhost(domain)
	if err != nil {
		c.Error(err)
		return
	}
	c.Set("data", data)
}

func (app *App) ApiAddQuicVhost(c *gin.Context) {
	data := &config.QuicVhost{}
	if err := c.ShouldBindJSON(data); err != nil {
		c.Error(err)
		return
	}

	app.configLock.Lock()
	defer app.configLock.Unlock()

	if err := app.DaoAddQuicVhost(data); err != nil {
		c.Error(err)
		return
	}
}

func (app *App) ApiDelQuicVhost(c *gin.Context) {
	domain := c.Param("domain")

	app.configLock.Lock()
	defer app.configLock.Unlock()

	if err := app.DaoRemoveQuicVhost(domain); err != nil {
		c.Error(err)
		return
	}
}

func (app *App) ApiModQuicVhost(c *gin.Context) {
	data := &config.QuicVhost{}
	if err := c.ShouldBindJSON(data); err != nil {
		c.Error(err)
		return
	}

	app.configLock.Lock()
	defer app.configLock.Unlock()

	if err := app.DaoSetQuicVhost(data); err != nil {
		c.Error(err)
		return
	}
}
