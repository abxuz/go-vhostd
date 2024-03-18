package middleware

import (
	"net/http"
	"strconv"
	"sync"

	"github.com/abxuz/b-tools/bstate"
	"github.com/abxuz/go-vhostd/internal/model"
	"github.com/gin-gonic/gin"
)

const DefaultAuthRealm = "Authorization Required"

func Auth(state *bstate.State[*model.AuthCfg], realm string) gin.HandlerFunc {
	var (
		lock    = new(sync.RWMutex)
		authCfg = state.Get()
	)

	state.Watch("middleware.Auth", func(prev, cur *model.AuthCfg) {
		lock.Lock()
		authCfg = cur
		lock.Unlock()
	})

	if realm == "" {
		realm = DefaultAuthRealm
	}
	realm = "Basic realm=" + strconv.Quote(realm)

	return func(ctx *gin.Context) {
		lock.RLock()
		cfg := authCfg
		lock.RUnlock()

		if cfg == nil {
			ctx.Next()
			return
		}

		username, password, ok := ctx.Request.BasicAuth()
		if !ok || username != cfg.Username || password != cfg.Password {
			ctx.Header("WWW-Authenticate", realm)
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		ctx.Next()
	}
}
