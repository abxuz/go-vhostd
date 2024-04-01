package api

import (
	"slices"
	"strings"

	"github.com/abxuz/b-tools/bslice"
	"github.com/abxuz/go-vhostd/internal/model"
	"github.com/abxuz/go-vhostd/internal/service"
	"github.com/gin-gonic/gin"
)

var Https = &aHttps{}

type aHttps struct {
}

func (a *aHttps) ListVhost() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		service.Cfg.MemoryLock(true)
		defer service.Cfg.MemoryUnlock(true)

		cfg, _ := service.Cfg.LoadFromMemory()

		vhost := make([]*model.HttpsVhostCfg, len(cfg.Https.Vhost))
		copy(vhost, cfg.Https.Vhost)

		slices.SortStableFunc(vhost, func(a, b *model.HttpsVhostCfg) int {
			return strings.Compare(a.Name, b.Name)
		})

		ctx.Set("resp", model.NewApiResponse(0).SetData(vhost))
	}
}

func (a *aHttps) GetVhost() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		domain := ctx.Param("domain")

		service.Cfg.MemoryLock(true)
		defer service.Cfg.MemoryUnlock(true)

		cfg, _ := service.Cfg.LoadFromMemory()
		i := bslice.FindIndex(cfg.Https.Vhost,
			func(c *model.HttpsVhostCfg) bool {
				return c.Domain == domain
			},
		)

		if i == -1 {
			ctx.Set("resp", model.NewApiResponse(1).SetErrMsg("vhost not found"))
			return
		}
		ctx.Set("resp", model.NewApiResponse(0).SetData(cfg.Https.Vhost[i]))
	}
}

func (a *aHttps) AddVhost() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req model.HttpsVhostCfg
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.Set("resp", model.NewApiResponse(1).SetErr(err))
			return
		}

		service.Cfg.MemoryLock(false)
		defer service.Cfg.MemoryUnlock(false)

		cfg, _ := service.Cfg.LoadFromMemory()
		cfg.Https.Vhost = append(cfg.Https.Vhost, &req)
		if err := service.Cfg.SaveToMemory(cfg); err != nil {
			ctx.Set("resp", model.NewApiResponse(1).SetErr(err))
			return
		}
		ctx.Set("resp", model.NewApiResponse(0))
	}
}

func (a *aHttps) ModVhost() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req model.HttpsVhostCfg
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.Set("resp", model.NewApiResponse(1).SetErr(err))
			return
		}

		service.Cfg.MemoryLock(false)
		defer service.Cfg.MemoryUnlock(false)

		cfg, _ := service.Cfg.LoadFromMemory()
		i := bslice.FindIndex(cfg.Https.Vhost,
			func(c *model.HttpsVhostCfg) bool {
				return c.Domain == req.Domain
			},
		)

		if i == -1 {
			ctx.Set("resp", model.NewApiResponse(1).SetErrMsg("vhost not found"))
			return
		}
		*cfg.Https.Vhost[i] = req
		if err := service.Cfg.SaveToMemory(cfg); err != nil {
			ctx.Set("resp", model.NewApiResponse(1).SetErr(err))
			return
		}
		ctx.Set("resp", model.NewApiResponse(0))
	}
}

func (a *aHttps) DelVhost() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		domain := ctx.Param("domain")

		service.Cfg.MemoryLock(false)
		defer service.Cfg.MemoryUnlock(false)

		cfg, _ := service.Cfg.LoadFromMemory()

		i := bslice.FindIndex(cfg.Https.Vhost,
			func(v *model.HttpsVhostCfg) bool {
				return v.Domain == domain
			},
		)
		if i == -1 {
			ctx.Set("resp", model.NewApiResponse(1).SetErrMsg("vhost not found"))
			return
		}
		cfg.Https.Vhost = append(cfg.Https.Vhost[:i], cfg.Https.Vhost[i+1:]...)

		if err := service.Cfg.SaveToMemory(cfg); err != nil {
			ctx.Set("resp", model.NewApiResponse(1).SetErr(err))
			return
		}
		ctx.Set("resp", model.NewApiResponse(0))
	}
}
