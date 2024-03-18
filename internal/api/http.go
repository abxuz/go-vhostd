package api

import (
	"github.com/abxuz/b-tools/bslice"
	"github.com/abxuz/go-vhostd/internal/model"
	"github.com/abxuz/go-vhostd/internal/service"
	"github.com/gin-gonic/gin"
)

var Http = &aHttp{}

type aHttp struct {
}

func (a *aHttp) GetConfig() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		service.Cfg.MemoryLock(true)
		defer service.Cfg.MemoryUnlock(true)

		cfg, _ := service.Cfg.LoadFromMemory()
		ctx.Set("resp", model.NewApiResponse(0).SetData(cfg.Http))
	}
}

func (a *aHttp) ListVhost() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		service.Cfg.MemoryLock(true)
		defer service.Cfg.MemoryUnlock(true)

		cfg, _ := service.Cfg.LoadFromMemory()
		ctx.Set("resp", model.NewApiResponse(0).SetData(cfg.Http.Vhost))
	}
}

func (a *aHttp) GetVhost() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		domain := ctx.Param("domain")

		service.Cfg.MemoryLock(true)
		defer service.Cfg.MemoryUnlock(true)

		cfg, _ := service.Cfg.LoadFromMemory()
		i := bslice.FindIndex(cfg.Http.Vhost,
			func(c *model.HttpVhostCfg) bool {
				return c.Domain == domain
			},
		)

		if i == -1 {
			ctx.Set("resp", model.NewApiResponse(1).SetErrMsg("vhost not found"))
			return
		}
		ctx.Set("resp", model.NewApiResponse(0).SetData(cfg.Http.Vhost[i]))
	}
}

func (a *aHttp) AddVhost() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req model.HttpVhostCfg
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.Set("resp", model.NewApiResponse(1).SetErr(err))
			return
		}

		service.Cfg.MemoryLock(false)
		defer service.Cfg.MemoryUnlock(false)

		cfg, _ := service.Cfg.LoadFromMemory()
		cfg.Http.Vhost = append(cfg.Http.Vhost, &req)
		if err := service.Cfg.SaveToMemory(cfg); err != nil {
			ctx.Set("resp", model.NewApiResponse(1).SetErr(err))
			return
		}
		ctx.Set("resp", model.NewApiResponse(0))
	}
}

func (a *aHttp) ModVhost() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req model.HttpVhostCfg
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.Set("resp", model.NewApiResponse(1).SetErr(err))
			return
		}

		service.Cfg.MemoryLock(false)
		defer service.Cfg.MemoryUnlock(false)

		cfg, _ := service.Cfg.LoadFromMemory()
		i := bslice.FindIndex(cfg.Http.Vhost,
			func(c *model.HttpVhostCfg) bool {
				return c.Domain == req.Domain
			},
		)

		if i == -1 {
			ctx.Set("resp", model.NewApiResponse(1).SetErrMsg("vhost not found"))
			return
		}
		*cfg.Http.Vhost[i] = req
		if err := service.Cfg.SaveToMemory(cfg); err != nil {
			ctx.Set("resp", model.NewApiResponse(1).SetErr(err))
			return
		}
		ctx.Set("resp", model.NewApiResponse(0))
	}
}

func (a *aHttp) DelVhost() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		domain := ctx.Param("domain")

		service.Cfg.MemoryLock(false)
		defer service.Cfg.MemoryUnlock(false)

		cfg, _ := service.Cfg.LoadFromMemory()

		i := bslice.FindIndex(cfg.Http.Vhost,
			func(v *model.HttpVhostCfg) bool {
				return v.Domain == domain
			},
		)
		if i == -1 {
			ctx.Set("resp", model.NewApiResponse(1).SetErrMsg("vhost not found"))
			return
		}
		cfg.Http.Vhost = append(cfg.Http.Vhost[:i], cfg.Http.Vhost[i+1:]...)

		if err := service.Cfg.SaveToMemory(cfg); err != nil {
			ctx.Set("resp", model.NewApiResponse(1).SetErr(err))
			return
		}
		ctx.Set("resp", model.NewApiResponse(0))
	}
}
