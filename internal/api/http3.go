package api

import (
	"github.com/abxuz/b-tools/bslice"
	"github.com/abxuz/go-vhostd/internal/model"
	"github.com/abxuz/go-vhostd/internal/service"
	"github.com/gin-gonic/gin"
)

var Http3 = &aHttp3{}

type aHttp3 struct {
}

func (a *aHttp3) GetConfig() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		service.Cfg.MemoryLock(true)
		defer service.Cfg.MemoryUnlock(true)

		cfg, _ := service.Cfg.LoadFromMemory()
		ctx.Set("resp", model.NewApiResponse(0).SetData(cfg.Http3))
	}
}

func (a *aHttp3) ListVhost() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		service.Cfg.MemoryLock(true)
		defer service.Cfg.MemoryUnlock(true)

		cfg, _ := service.Cfg.LoadFromMemory()
		ctx.Set("resp", model.NewApiResponse(0).SetData(cfg.Http3.Vhost))
	}
}

func (a *aHttp3) GetVhost() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		domain := ctx.Param("domain")

		service.Cfg.MemoryLock(true)
		defer service.Cfg.MemoryUnlock(true)

		cfg, _ := service.Cfg.LoadFromMemory()
		i := bslice.FindIndex(cfg.Http3.Vhost,
			func(c *model.Http3VhostCfg) bool {
				return c.Domain == domain
			},
		)

		if i == -1 {
			ctx.Set("resp", model.NewApiResponse(1).SetErrMsg("vhost not found"))
			return
		}
		ctx.Set("resp", model.NewApiResponse(0).SetData(cfg.Http3.Vhost[i]))
	}
}

func (a *aHttp3) AddVhost() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req model.Http3VhostCfg
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.Set("resp", model.NewApiResponse(1).SetErr(err))
			return
		}

		service.Cfg.MemoryLock(false)
		defer service.Cfg.MemoryUnlock(false)

		cfg, _ := service.Cfg.LoadFromMemory()
		cfg.Http3.Vhost = append(cfg.Http3.Vhost, &req)
		if err := service.Cfg.SaveToMemory(cfg); err != nil {
			ctx.Set("resp", model.NewApiResponse(1).SetErr(err))
			return
		}
		ctx.Set("resp", model.NewApiResponse(0))
	}
}

func (a *aHttp3) ModVhost() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req model.Http3VhostCfg
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.Set("resp", model.NewApiResponse(1).SetErr(err))
			return
		}

		service.Cfg.MemoryLock(false)
		defer service.Cfg.MemoryUnlock(false)

		cfg, _ := service.Cfg.LoadFromMemory()
		i := bslice.FindIndex(cfg.Http3.Vhost,
			func(c *model.Http3VhostCfg) bool {
				return c.Domain == req.Domain
			},
		)

		if i == -1 {
			ctx.Set("resp", model.NewApiResponse(1).SetErrMsg("vhost not found"))
			return
		}
		*cfg.Http3.Vhost[i] = req
		if err := service.Cfg.SaveToMemory(cfg); err != nil {
			ctx.Set("resp", model.NewApiResponse(1).SetErr(err))
			return
		}
		ctx.Set("resp", model.NewApiResponse(0))
	}
}

func (a *aHttp3) DelVhost() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		domain := ctx.Param("domain")

		service.Cfg.MemoryLock(false)
		defer service.Cfg.MemoryUnlock(false)

		cfg, _ := service.Cfg.LoadFromMemory()

		i := bslice.FindIndex(cfg.Http3.Vhost,
			func(v *model.Http3VhostCfg) bool {
				return v.Domain == domain
			},
		)
		if i == -1 {
			ctx.Set("resp", model.NewApiResponse(1).SetErrMsg("vhost not found"))
			return
		}
		cfg.Http3.Vhost = append(cfg.Http3.Vhost[:i], cfg.Http3.Vhost[i+1:]...)

		if err := service.Cfg.SaveToMemory(cfg); err != nil {
			ctx.Set("resp", model.NewApiResponse(1).SetErr(err))
			return
		}
		ctx.Set("resp", model.NewApiResponse(0))
	}
}
