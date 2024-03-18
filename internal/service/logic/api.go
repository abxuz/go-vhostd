package logic

import (
	"io"
	"log"
	"net/http"

	"github.com/abxuz/b-tools/bhttp"
	"github.com/abxuz/b-tools/bset"
	"github.com/abxuz/b-tools/bstate"
	"github.com/abxuz/go-vhostd/assets"
	"github.com/abxuz/go-vhostd/internal/api"
	"github.com/abxuz/go-vhostd/internal/middleware"
	"github.com/abxuz/go-vhostd/internal/model"
	"github.com/abxuz/go-vhostd/internal/service"
	"github.com/gin-gonic/gin"
)

type lApi struct {
	servers   map[string]*http.Server
	authState *bstate.State[*model.AuthCfg]
	handler   *gin.Engine
}

func init() {
	service.RegisterApiService(&lApi{})
}

func (l *lApi) Init() {
	l.servers = make(map[string]*http.Server)
	l.authState = bstate.NewState[*model.AuthCfg]()

	gin.SetMode(gin.ReleaseMode)
	l.handler = gin.New()
	l.handler.Use(
		gin.RecoveryWithWriter(io.Discard),
		middleware.Auth(l.authState, middleware.DefaultAuthRealm),
	)

	fs := http.FileServer(
		&bhttp.NoAutoIndexFileSystem{
			FileSystem: http.FS(assets.Html),
		},
	)
	l.handler.NoRoute(gin.WrapH(fs))

	v1 := l.handler.Group("/api/v1/")
	v1.Use(middleware.ApiResponse())
	{
		v1.GET("/reload", api.Api.Reload())
		v1.GET("/save", api.Api.Save())

		v1.GET("/vhost-listen", api.Api.GetVhostListen())
		v1.POST("/vhost-listen", api.Api.SetVhostListen())

		v1.GET("/api-config", api.Api.GetApiConfig())
		v1.POST("/api-config", api.Api.SetApiConfig())

		v1.GET("/http-config", api.Http.GetConfig())
		g := v1.Group("/http-vhost/")
		{
			g.POST("/", api.Http.AddVhost())
			g.DELETE("/:domain", api.Http.DelVhost())
			g.PATCH("/", api.Http.ModVhost())
			g.GET("/", api.Http.ListVhost())
			g.GET("/:domain", api.Http.GetVhost())
		}

		v1.GET("/https-config", api.Https.GetConfig())
		g = v1.Group("/https-vhost/")
		{
			g.POST("/", api.Https.AddVhost())
			g.DELETE("/:domain", api.Https.DelVhost())
			g.PATCH("/", api.Https.ModVhost())
			g.GET("/", api.Https.ListVhost())
			g.GET("/:domain", api.Https.GetVhost())
		}

		v1.GET("/http3-config", api.Http3.GetConfig())
		g = v1.Group("/http3-vhost/")
		{
			g.POST("/", api.Http3.AddVhost())
			g.DELETE("/:domain", api.Http3.DelVhost())
			g.PATCH("/", api.Http3.ModVhost())
			g.GET("/", api.Http3.ListVhost())
			g.GET("/:domain", api.Http3.GetVhost())
		}

		g = v1.Group("/cert/")
		{
			g.POST("/", api.Cert.Add())
			g.DELETE("/:name", api.Cert.Del())
			g.PATCH("/", api.Cert.Mod())
			g.GET("/", api.Cert.List())
			g.GET("/:name", api.Cert.Get())
		}
	}
}

func (l *lApi) Reload(cfg model.Cfg) {
	l.authState.Set(cfg.Api.Auth)

	listen := bset.New(cfg.Api.Listen...)
	for k, server := range l.servers {
		if listen.Has(k) {
			listen.Remove(k)
			continue
		}
		server.Close()
		delete(l.servers, k)
	}

	listen.Range(func(k string) bool {
		server := &http.Server{
			Addr:     k,
			Handler:  l.handler,
			ErrorLog: log.New(io.Discard, "", log.LstdFlags),
		}
		go server.ListenAndServe()
		l.servers[k] = server
		return true
	})
}
