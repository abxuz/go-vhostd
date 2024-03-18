package service

import (
	"github.com/abxuz/go-vhostd/internal/model"
)

type ApiService interface {
	Init()
	Reload(cfg model.Cfg)
}

var Api ApiService

func RegisterApiService(s ApiService) {
	Api = s
}
