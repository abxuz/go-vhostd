package app

import (
	iofs "io/fs"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/quic-go/quic-go/http3"
	"github.com/xbugio/go-vhostd/internal/fs"
)

type App struct {
	wg               *sync.WaitGroup
	embedFileHandler http.Handler

	config     string
	configLock *sync.RWMutex

	apiState     *ApiState
	apiStateLock *sync.RWMutex
	apiServer    *http.Server

	httpState     *HttpState
	httpStateLock *sync.RWMutex
	httpServer    *http.Server

	httpsState     *HttpsState
	httpsStateLock *sync.RWMutex
	httpsServer    *http.Server

	quicState     *QuicState
	quicStateLock *sync.RWMutex
	quicServer    *http3.Server

	certState     *CertState
	certStateLock *sync.RWMutex
}

func NewApp(c string, embedFs iofs.FS) *App {
	return &App{
		wg: &sync.WaitGroup{},
		embedFileHandler: http.FileServer(&fs.NoAutoIndexFileSystem{
			FileSystem: http.FS(embedFs),
		}),

		config:     c,
		configLock: &sync.RWMutex{},

		apiState:     nil,
		apiStateLock: &sync.RWMutex{},
		apiServer:    nil,

		httpState:     nil,
		httpStateLock: &sync.RWMutex{},
		httpServer:    nil,

		httpsState:     nil,
		httpsStateLock: &sync.RWMutex{},
		httpsServer:    nil,

		quicState:     nil,
		quicStateLock: &sync.RWMutex{},
		quicServer:    nil,

		certState:     nil,
		certStateLock: &sync.RWMutex{},
	}
}

func (app *App) Run() error {
	gin.SetMode(gin.ReleaseMode)

	if err := app.ServiceReload(); err != nil {
		return err
	}

	app.wg.Wait()
	return nil
}
