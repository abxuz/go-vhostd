package service

import "github.com/abxuz/go-vhostd/internal/model"

type ProxyService interface {
	Init()
	Reload(cfg model.Cfg)
}

var Proxy ProxyService

func RegisterProxyService(s ProxyService) {
	Proxy = s
}
