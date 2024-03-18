package service

import "github.com/abxuz/go-vhostd/internal/model"

type CfgService interface {
	SetFilePath(config string, init bool)

	FileLock(readonly bool)
	FileUnlock(readonly bool)
	LoadFromFile() (model.Cfg, error)
	SaveToFile(cfg model.Cfg) error

	MemoryLock(readonly bool)
	MemoryUnlock(readonly bool)
	LoadFromMemory() (model.Cfg, error)
	SaveToMemory(cfg model.Cfg) error
}

var Cfg CfgService

func RegisterCfgService(s CfgService) {
	Cfg = s
}
