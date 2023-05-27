package app

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xbugio/go-vhostd/internal/config"
)

func (app *App) newApiServer(state *ApiState) *http.Server {
	router := gin.New()
	router.Use(gin.RecoveryWithWriter(io.Discard))
	router.Use(app.BasicAuthMiddleware)

	v1 := router.Group("/api/v1/")
	v1.Use(app.ApiResponseMiddleware)
	{
		v1.GET("/reload", app.ApiReload)

		v1.GET("/config", app.ApiGetConfig)
		v1.POST("/config", app.ApiSetConfig)

		v1.GET("/vhost-listen", app.ApiGetVhostListen)
		v1.POST("/vhost-listen", app.ApiSetVhostListen)

		v1.GET("/api-config", app.ApiGetApiConfig)
		v1.POST("/api-config", app.ApiSetApiConfig)

		v1.GET("/http-config", app.ApiGetHttpConfig)
		v1.POST("/http-config", app.ApiSetHttpConfig)
		g := v1.Group("/http-vhost/")
		{
			g.POST("/", app.ApiAddHttpVhost)
			g.DELETE("/:domain", app.ApiDelHttpVhost)
			g.PATCH("/", app.ApiModHttpVhost)
			g.GET("/", app.ApiListHttpVhost)
			g.GET("/:domain", app.ApiGetHttpVhost)
		}

		v1.GET("/https-config", app.ApiGetHttpsConfig)
		v1.POST("/https-config", app.ApiSetHttpsConfig)
		g = v1.Group("/https-vhost/")
		{
			g.POST("/", app.ApiAddHttpsVhost)
			g.DELETE("/:domain", app.ApiDelHttpsVhost)
			g.PATCH("/", app.ApiModHttpsVhost)
			g.GET("/", app.ApiListHttpsVhost)
			g.GET("/:domain", app.ApiGetHttpsVhost)
		}

		v1.GET("/quic-config", app.ApiGetQuicConfig)
		v1.POST("/quic-config", app.ApiSetQuicConfig)
		g = v1.Group("/quic-vhost/")
		{
			g.POST("/", app.ApiAddQuicVhost)
			g.DELETE("/:domain", app.ApiDelQuicVhost)
			g.PATCH("/", app.ApiModQuicVhost)
			g.GET("/", app.ApiListQuicVhost)
			g.GET("/:domain", app.ApiGetQuicVhost)
		}

		v1.GET("/cert-config", app.ApiGetCertConfig)
		v1.POST("/cert-config", app.ApiSetCertConfig)
		g = v1.Group("/cert/")
		{
			g.POST("/", app.ApiAddCert)
			g.DELETE("/:name", app.ApiDelCert)
			g.PATCH("/", app.ApiModCert)
			g.GET("/", app.ApiListCert)
			g.GET("/:name", app.ApiGetCert)
		}
	}

	router.NoRoute(app.ServeFrontendFile)

	return &http.Server{
		Addr:     state.Listen,
		Handler:  router,
		ErrorLog: LogDiscard,
	}
}

func (app *App) ServeFrontendFile(c *gin.Context) {
	app.apiStateLock.RLock()
	defer app.apiStateLock.RUnlock()

	fileHandler := app.frontendFileHandler
	if app.apiState.PublicFileHandler != nil {
		fileHandler = app.apiState.PublicFileHandler
	}
	fileHandler.ServeHTTP(c.Writer, c.Request)
}

func (app *App) BasicAuthMiddleware(c *gin.Context) {
	app.apiStateLock.RLock()

	if app.apiState.Auth == nil {
		app.apiStateLock.RUnlock()
		c.Next()
		return
	}

	username, password, ok := c.Request.BasicAuth()
	authOk := ok && (username == app.apiState.Auth.Username) && (password == app.apiState.Auth.Password)
	app.apiStateLock.RUnlock()

	if authOk {
		c.Next()
		return
	}

	c.Header("WWW-Authenticate", "Basic realm=Authorization Required")
	c.AbortWithStatus(http.StatusUnauthorized)
}

func (app *App) ApiResponseMiddleware(c *gin.Context) {
	c.Next()

	err := c.Errors.Last()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"errno":  1,
			"errmsg": err.Error(),
		})
		return
	}

	obj := gin.H{
		"errno": 0,
	}
	data, exists := c.Get("data")
	if exists {
		obj["data"] = data
	}
	c.JSON(http.StatusOK, obj)
}

func (app *App) ApiReload(c *gin.Context) {
	app.ServiceReload()
}

func (app *App) ApiGetConfig(c *gin.Context) {
	app.configLock.RLock()
	defer app.configLock.RUnlock()
	cfg, err := app.DaoGetCfg()
	if err != nil {
		c.Error(err)
		return
	}
	c.Set("data", cfg)
}

func (app *App) ApiSetConfig(c *gin.Context) {
	cfg := &config.Config{}
	if err := c.ShouldBindJSON(cfg); err != nil {
		c.Error(err)
		return
	}

	app.configLock.Lock()
	defer app.configLock.Unlock()

	if err := app.DaoSetCfg(cfg); err != nil {
		c.Error(err)
		return
	}
}

func (app *App) ApiGetApiConfig(c *gin.Context) {
	app.configLock.RLock()
	defer app.configLock.RUnlock()
	cfg, err := app.DaoGetCfg()
	if err != nil {
		c.Error(err)
		return
	}
	c.Set("data", cfg.Api)
}

func (app *App) ApiSetApiConfig(c *gin.Context) {
	apiCfg := &config.Api{}
	if err := c.ShouldBindJSON(apiCfg); err != nil {
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
	cfg.Api = apiCfg
	if err := app.DaoSetCfg(cfg); err != nil {
		c.Error(err)
		return
	}
}

func (app *App) ApiGetVhostListen(c *gin.Context) {
	app.configLock.RLock()
	defer app.configLock.RUnlock()

	cfg, err := app.DaoGetCfg()
	if err != nil {
		c.Error(err)
		return
	}

	data := &VhostListen{}
	if cfg.Http != nil {
		data.Http = cfg.Http.Listen
	}

	if cfg.Https != nil {
		data.Https = cfg.Https.Listen
	}

	if cfg.Quic != nil {
		data.Quic = cfg.Quic.Listen
	}

	c.Set("data", data)
}

func (app *App) ApiSetVhostListen(c *gin.Context) {
	data := &VhostListen{}
	if err := c.ShouldBindJSON(data); err != nil {
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

	if cfg.Http == nil {
		cfg.Http = &config.Http{}
	}
	if cfg.Https == nil {
		cfg.Https = &config.Https{}
	}
	if cfg.Quic == nil {
		cfg.Quic = &config.Quic{}
	}

	cfg.Http.Listen = data.Http
	cfg.Https.Listen = data.Https
	cfg.Quic.Listen = data.Quic

	if err := app.DaoSetCfg(cfg); err != nil {
		c.Error(err)
		return
	}
}
