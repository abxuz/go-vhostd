package app

import (
	"github.com/gin-gonic/gin"
	"github.com/xbugio/go-vhostd/internal/config"
)

func (app *App) ApiGetCertConfig(c *gin.Context) {
	app.configLock.RLock()
	defer app.configLock.RUnlock()
	cfg, err := app.DaoGetCfg()
	if err != nil {
		c.Error(err)
		return
	}
	c.Set("data", cfg.Cert)
}

func (app *App) ApiSetCertConfig(c *gin.Context) {
	list := make([]*config.Cert, 0)
	if err := c.ShouldBindJSON(list); err != nil {
		c.Error(err)
		return
	}

	app.configLock.Lock()
	defer app.configLock.Unlock()

	err := app.DaoSetCerts(list)
	if err != nil {
		c.Error(err)
		return
	}
}

func (app *App) ApiListCert(c *gin.Context) {
	app.configLock.RLock()
	defer app.configLock.RUnlock()
	list, err := app.DaoGetCerts()
	if err != nil {
		c.Error(err)
		return
	}

	type dataType struct {
		*config.Cert
		*CertInfo
	}

	data := make([]*dataType, 0)
	for _, certCfg := range list {
		info, err := app.DaoGetCertInfo(certCfg)
		if err != nil {
			c.Error(err)
			return
		}
		data = append(data, &dataType{
			Cert:     certCfg,
			CertInfo: info,
		})
	}
	c.Set("data", data)
}

func (app *App) ApiGetCert(c *gin.Context) {
	name := c.Param("name")

	app.configLock.RLock()
	defer app.configLock.RUnlock()

	certCfg, err := app.DaoGetCert(name)
	if err != nil {
		c.Error(err)
		return
	}

	info, err := app.DaoGetCertInfo(certCfg)
	if err != nil {
		c.Error(err)
		return
	}

	type dataType struct {
		*config.Cert
		*CertInfo
	}

	c.Set("data", &dataType{
		Cert:     certCfg,
		CertInfo: info,
	})
}

func (app *App) ApiAddCert(c *gin.Context) {
	cert := &config.Cert{}
	if err := c.ShouldBindJSON(cert); err != nil {
		c.Error(err)
		return
	}

	app.configLock.Lock()
	defer app.configLock.Unlock()

	if err := app.DaoAddCert(cert); err != nil {
		c.Error(err)
		return
	}
}

func (app *App) ApiDelCert(c *gin.Context) {
	name := c.Param("name")

	app.configLock.Lock()
	defer app.configLock.Unlock()

	if err := app.DaoRemoveCert(name); err != nil {
		c.Error(err)
		return
	}
}

func (app *App) ApiModCert(c *gin.Context) {
	cert := &config.Cert{}
	if err := c.ShouldBindJSON(cert); err != nil {
		c.Error(err)
		return
	}

	app.configLock.Lock()
	defer app.configLock.Unlock()

	if err := app.DaoSetCert(cert); err != nil {
		c.Error(err)
		return
	}
}
