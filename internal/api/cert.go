package api

import (
	"github.com/abxuz/b-tools/bslice"
	"github.com/abxuz/go-vhostd/internal/model"
	"github.com/abxuz/go-vhostd/internal/service"
	"github.com/gin-gonic/gin"
)

var Cert = &aCert{}

type aCert struct {
}

type CertResponse struct {
	*model.CertCfg
	*model.CertInfo
}

func (a *aCert) List() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		service.Cfg.MemoryLock(true)
		defer service.Cfg.MemoryUnlock(true)

		cfg, _ := service.Cfg.LoadFromMemory()

		list := make([]*CertResponse, 0)
		for _, c := range cfg.Cert {
			info, err := c.CertInfo()
			if err != nil {
				ctx.Set("resp", model.NewApiResponse(1).SetErr(err))
				return
			}
			list = append(list, &CertResponse{
				CertCfg:  c,
				CertInfo: info,
			})
		}

		ctx.Set("resp", model.NewApiResponse(0).SetData(list))
	}
}

func (a *aCert) Get() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		name := ctx.Param("name")

		service.Cfg.MemoryLock(true)
		defer service.Cfg.MemoryUnlock(true)

		cfg, _ := service.Cfg.LoadFromMemory()
		i := bslice.FindIndex(cfg.Cert,
			func(c *model.CertCfg) bool {
				return c.Name == name
			},
		)

		if i == -1 {
			ctx.Set("resp", model.NewApiResponse(1).SetErrMsg("cert not found"))
			return
		}

		cert := cfg.Cert[i]
		info, err := cert.CertInfo()
		if err != nil {
			ctx.Set("resp", model.NewApiResponse(1).SetErr(err))
			return
		}

		resp := &CertResponse{
			CertCfg:  cert,
			CertInfo: info,
		}
		ctx.Set("resp", model.NewApiResponse(0).SetData(resp))
	}
}

func (a *aCert) Add() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req model.CertCfg
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.Set("resp", model.NewApiResponse(1).SetErr(err))
			return
		}

		service.Cfg.MemoryLock(false)
		defer service.Cfg.MemoryUnlock(false)

		cfg, _ := service.Cfg.LoadFromMemory()
		cfg.Cert = append(cfg.Cert, &req)

		if err := service.Cfg.SaveToMemory(cfg); err != nil {
			ctx.Set("resp", model.NewApiResponse(1).SetErr(err))
			return
		}
		ctx.Set("resp", model.NewApiResponse(0))
	}
}

func (a *aCert) Mod() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req model.CertCfg
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.Set("resp", model.NewApiResponse(1).SetErr(err))
			return
		}

		service.Cfg.MemoryLock(false)
		defer service.Cfg.MemoryUnlock(false)

		cfg, _ := service.Cfg.LoadFromMemory()
		i := bslice.FindIndex(cfg.Cert,
			func(c *model.CertCfg) bool {
				return c.Name == req.Name
			},
		)

		if i == -1 {
			ctx.Set("resp", model.NewApiResponse(1).SetErrMsg("cert not found"))
			return
		}

		*cfg.Cert[i] = req
		if err := service.Cfg.SaveToMemory(cfg); err != nil {
			ctx.Set("resp", model.NewApiResponse(1).SetErr(err))
			return
		}
		ctx.Set("resp", model.NewApiResponse(0))
	}
}

func (a *aCert) Del() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		name := ctx.Param("name")

		service.Cfg.MemoryLock(false)
		defer service.Cfg.MemoryUnlock(false)

		cfg, _ := service.Cfg.LoadFromMemory()
		for _, vhost := range cfg.Https.Vhost {
			if vhost.Cert == name {
				ctx.Set("resp", model.NewApiResponse(1).SetErrMsg("cert is in use"))
				return
			}
		}

		for _, vhost := range cfg.Http3.Vhost {
			if vhost.Cert == name {
				ctx.Set("resp", model.NewApiResponse(1).SetErrMsg("cert is in use"))
				return
			}
		}

		i := bslice.FindIndex(cfg.Cert,
			func(c *model.CertCfg) bool {
				return c.Name == name
			},
		)

		if i == -1 {
			ctx.Set("resp", model.NewApiResponse(1).SetErrMsg("cert not found"))
			return
		}
		cfg.Cert = append(cfg.Cert[:i], cfg.Cert[i+1:]...)

		if err := service.Cfg.SaveToMemory(cfg); err != nil {
			ctx.Set("resp", model.NewApiResponse(1).SetErr(err))
			return
		}
		ctx.Set("resp", model.NewApiResponse(0))
	}
}
