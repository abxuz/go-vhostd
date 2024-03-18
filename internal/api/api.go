package api

import (
	"github.com/abxuz/go-vhostd/internal/model"
	"github.com/abxuz/go-vhostd/internal/service"
	"github.com/gin-gonic/gin"
)

var Api = &aApi{}

type aApi struct {
}

func (a *aApi) Reload() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		service.Cfg.MemoryLock(true)
		defer service.Cfg.MemoryUnlock(true)

		cfg, _ := service.Cfg.LoadFromMemory()
		service.Proxy.Reload(cfg)
		service.Api.Reload(cfg)

		ctx.Set("resp", model.NewApiResponse(0))
	}
}

func (a *aApi) Save() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		service.Cfg.MemoryLock(true)
		defer service.Cfg.MemoryUnlock(true)

		service.Cfg.FileLock(false)
		defer service.Cfg.FileUnlock(false)

		cfg, _ := service.Cfg.LoadFromMemory()
		if err := service.Cfg.SaveToFile(cfg); err != nil {
			ctx.Set("resp", model.NewApiResponse(1).SetErr(err))
			return
		}
		ctx.Set("resp", model.NewApiResponse(0))
	}
}

func (a *aApi) GetApiConfig() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		service.Cfg.MemoryLock(true)
		defer service.Cfg.MemoryUnlock(true)

		cfg, _ := service.Cfg.LoadFromMemory()
		ctx.Set("resp", model.NewApiResponse(0).SetData(cfg.Api))
	}
}

func (a *aApi) SetApiConfig() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req model.ApiCfg
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.Set("resp", model.NewApiResponse(1).SetErr(err))
			return
		}

		service.Cfg.MemoryLock(false)
		defer service.Cfg.MemoryUnlock(false)

		cfg, _ := service.Cfg.LoadFromMemory()
		*cfg.Api = req
		if err := service.Cfg.SaveToMemory(cfg); err != nil {
			ctx.Set("resp", model.NewApiResponse(1).SetErr(err))
			return
		}
		ctx.Set("resp", model.NewApiResponse(0))
	}
}

type GetVhostListenResponse struct {
	Http  []string `json:"http"`
	Https []string `json:"https"`
	Http3 []string `json:"http3"`
}

func (a *aApi) GetVhostListen() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		service.Cfg.MemoryLock(true)
		defer service.Cfg.MemoryUnlock(true)

		cfg, _ := service.Cfg.LoadFromMemory()
		data := GetVhostListenResponse{
			Http:  cfg.Http.Listen,
			Https: cfg.Https.Listen,
			Http3: cfg.Http3.Listen,
		}
		ctx.Set("resp", model.NewApiResponse(0).SetData(data))
	}
}

type SetVhostListenRequest struct {
	Http  []string `json:"http"`
	Https []string `json:"https"`
	Http3 []string `json:"http3"`
}

func (a *aApi) SetVhostListen() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req SetVhostListenRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.Set("resp", model.NewApiResponse(1).SetErr(err))
			return
		}

		service.Cfg.MemoryLock(false)
		defer service.Cfg.MemoryUnlock(false)

		cfg, _ := service.Cfg.LoadFromMemory()
		cfg.Http.Listen = req.Http
		cfg.Https.Listen = req.Https
		cfg.Http3.Listen = req.Http3
		if err := service.Cfg.SaveToMemory(cfg); err != nil {
			ctx.Set("resp", model.NewApiResponse(1).SetErr(err))
			return
		}
		ctx.Set("resp", model.NewApiResponse(0))
	}
}
