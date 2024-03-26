package logic

import (
	"io"
	"os"
	"sync"

	"github.com/abxuz/go-vhostd/internal/model"
	"github.com/abxuz/go-vhostd/internal/service"
	"gopkg.in/yaml.v3"
)

type lCfg struct {
	file     string
	fileLock sync.RWMutex

	memCfg     model.Cfg
	memCfgLock sync.RWMutex
}

func init() {
	service.RegisterCfgService(&lCfg{})
}

func (l *lCfg) SetFilePath(config string, init bool) {
	l.file = config
	if !init {
		return
	}

	_, err := os.Stat(config)
	if err == nil {
		return
	}

	if !os.IsNotExist(err) {
		return
	}

	l.SaveToFile(model.Cfg{
		Api: &model.ApiCfg{
			Listen: []string{":80"},
		},
	})
}

func (l *lCfg) FileLock(readonly bool) {
	if readonly {
		l.fileLock.RLock()
	} else {
		l.fileLock.Lock()
	}
}

func (l *lCfg) FileUnlock(readonly bool) {
	if readonly {
		l.fileLock.RUnlock()
	} else {
		l.fileLock.Unlock()
	}
}

func (l *lCfg) LoadFromFile() (cfg model.Cfg, err error) {
	var file *os.File
	file, err = os.Open(l.file)
	if err != nil {
		return
	}
	defer file.Close()
	return l.decode(file)
}

func (l *lCfg) SaveToFile(cfg model.Cfg) error {
	file, err := os.Create(l.file)
	if err != nil {
		return err
	}
	defer file.Close()
	return l.encode(&cfg, file)
}

func (l *lCfg) MemoryLock(readonly bool) {
	if readonly {
		l.memCfgLock.RLock()
	} else {
		l.memCfgLock.Lock()
	}
}

func (l *lCfg) MemoryUnlock(readonly bool) {
	if readonly {
		l.memCfgLock.RUnlock()
	} else {
		l.memCfgLock.Unlock()
	}
}

func (l *lCfg) LoadFromMemory() (cfg model.Cfg, err error) {
	cfg = l.memCfg
	return
}

func (l *lCfg) SaveToMemory(cfg model.Cfg) error {
	l.autofill(&cfg)
	if err := cfg.CheckValid(); err != nil {
		return err
	}

	l.memCfg = cfg
	return nil
}

func (l *lCfg) decode(r io.Reader) (cfg model.Cfg, err error) {
	err = yaml.NewDecoder(r).Decode(&cfg)
	if err != nil {
		return
	}
	l.autofill(&cfg)
	err = cfg.CheckValid()
	return
}

func (l *lCfg) encode(cfg *model.Cfg, w io.Writer) error {
	l.autofill(cfg)
	if err := cfg.CheckValid(); err != nil {
		return err
	}
	encoder := yaml.NewEncoder(w)
	defer encoder.Close()
	return encoder.Encode(cfg)
}

func (l *lCfg) autofill(cfg *model.Cfg) {
	if cfg.Api == nil {
		cfg.Api = &model.ApiCfg{}
	}

	if cfg.Api.Listen == nil {
		cfg.Api.Listen = make([]string, 0)
	}

	if cfg.Http == nil {
		cfg.Http = &model.HttpCfg{}
	}
	if cfg.Http.Listen == nil {
		cfg.Http.Listen = make([]string, 0)
	}
	if cfg.Http.Vhost == nil {
		cfg.Http.Vhost = make([]*model.HttpVhostCfg, 0)
	}
	for _, vhost := range cfg.Http.Vhost {
		if vhost.Mapping == nil {
			vhost.Mapping = make([]*model.MappingCfg, 0)
		}
		for _, h := range vhost.Mapping {
			if h.AddHeader == nil {
				h.AddHeader = make([]string, 0)
			}
		}
	}

	if cfg.Https == nil {
		cfg.Https = &model.HttpsCfg{}
	}
	if cfg.Https.Listen == nil {
		cfg.Https.Listen = make([]string, 0)
	}
	if cfg.Https.Vhost == nil {
		cfg.Https.Vhost = make([]*model.HttpsVhostCfg, 0)
	}
	for _, vhost := range cfg.Https.Vhost {
		if vhost.Mapping == nil {
			vhost.Mapping = make([]*model.MappingCfg, 0)
		}
		for _, h := range vhost.Mapping {
			if h.AddHeader == nil {
				h.AddHeader = make([]string, 0)
			}
		}
	}

	if cfg.Http3 == nil {
		cfg.Http3 = &model.Http3Cfg{}
	}
	if cfg.Http3.Listen == nil {
		cfg.Http3.Listen = make([]string, 0)
	}
	if cfg.Http3.Vhost == nil {
		cfg.Http3.Vhost = make([]*model.Http3VhostCfg, 0)
	}
	for _, vhost := range cfg.Http3.Vhost {
		if vhost.Mapping == nil {
			vhost.Mapping = make([]*model.MappingCfg, 0)
		}
		for _, h := range vhost.Mapping {
			if h.AddHeader == nil {
				h.AddHeader = make([]string, 0)
			}
		}
	}

	if cfg.Cert == nil {
		cfg.Cert = make([]*model.CertCfg, 0)
	}
}
