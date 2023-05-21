package app

import (
	"github.com/xbugio/go-vhostd/internal/config"
)

func (app *App) ServiceReload() error {
	app.configLock.RLock()
	defer app.configLock.RUnlock()

	cfg, err := config.LoadConfigFromFile(app.config)
	if err != nil {
		return err
	}

	state, err := GenState(cfg)
	if err != nil {
		return err
	}

	if err := app.ServiceReloadApiServer(state); err != nil {
		return err
	}

	if err := app.ServiceReloadHttpVhostServer(state); err != nil {
		return err
	}

	if err := app.ServiceReloadHttpsVhostServer(state); err != nil {
		return err
	}

	if err := app.ServiceReloadQuicVhostServer(state); err != nil {
		return err
	}

	if err := app.ServiceReloadCert(state); err != nil {
		return err
	}

	return nil
}

func (app *App) ServiceReloadApiServer(state *State) error {
	app.apiStateLock.Lock()
	defer app.apiStateLock.Unlock()

	// 新配置为关闭api server
	if state.Api == nil || state.Api.Listen == "" {
		// 老api server在运行，就关闭
		if app.apiState != nil && app.apiState.Listen != "" {
			app.apiServer.Close()
			app.apiServer = nil
		}
		app.apiState = state.Api
		return nil
	}

	// 否则新配置为要运行api server
	// 判断老的api server是否在运行
	// 在运行
	if app.apiState != nil && app.apiState.Listen != "" {

		// 判断监听地址跟新的是否一样
		// 一样就不动老的了
		if app.apiState.Listen == state.Api.Listen {
			app.apiState = state.Api
			return nil
		}

		// 不一样，就把老的停掉
		app.apiServer.Close()
		app.apiServer = nil
	}

	// 老的没在运行了
	// 建个新的
	// 新的没跑起来也没办法了，老的没在跑了
	app.apiState = state.Api
	app.apiServer = app.newApiServer(state.Api)
	app.wg.Add(1)
	go func(app *App) {
		defer app.wg.Done()
		app.apiServer.ListenAndServe()
	}(app)
	return nil
}

func (app *App) ServiceReloadHttpVhostServer(state *State) error {
	app.httpStateLock.Lock()
	defer app.httpStateLock.Unlock()

	// 新配置为关闭vhost server
	if state.Http == nil || state.Http.Listen == "" {
		// 老vhost server在运行，就关闭
		if app.httpState != nil && app.httpState.Listen != "" {
			app.httpServer.Close()
			app.httpServer = nil
		}
		app.httpState = state.Http
		return nil
	}

	// 否则配置为要运行vhost server
	// 判断老的vhost server是否在运行
	// 在运行
	if app.httpState != nil && app.httpState.Listen != "" {
		// 判断监听地址跟新的是否一样
		// 一样就不动老的了
		if app.httpState.Listen == state.Http.Listen {
			app.httpState = state.Http
			return nil
		}

		// 不一样，就把老的停掉
		app.httpServer.Close()
		app.httpServer = nil
	}

	// 老的没在运行了
	// 建个新的
	// 新的没跑起来也没办法了，老的没在跑了
	app.httpState = state.Http
	app.httpServer = app.newHttpServer(state.Http)
	app.wg.Add(1)
	go func(app *App) {
		defer app.wg.Done()
		app.httpServer.ListenAndServe()
	}(app)
	return nil
}

func (app *App) ServiceReloadHttpsVhostServer(state *State) error {
	app.httpsStateLock.Lock()
	defer app.httpsStateLock.Unlock()

	// 新配置为关闭vhost server
	if state.Https == nil || state.Https.Listen == "" {
		// 老vhost server在运行，就关闭
		if app.httpsState != nil && app.httpsState.Listen != "" {
			app.httpsServer.Close()
			app.httpsServer = nil
		}
		app.httpsState = state.Https
		return nil
	}

	// 否则配置为要运行vhost server
	// 判断老的vhost server是否在运行
	// 在运行
	if app.httpsState != nil && app.httpsState.Listen != "" {
		// 判断监听地址跟新的是否一样
		// 一样就不动老的了
		if app.httpsState.Listen == state.Https.Listen {
			app.httpsState = state.Https
			return nil
		}

		// 不一样，就把老的停掉
		app.httpsServer.Close()
		app.httpsServer = nil
	}

	// 老的没在运行了
	// 建个新的
	// 新的没跑起来也没办法了，老的没在跑了
	app.httpsState = state.Https
	app.httpsServer = app.newHttpsServer(state.Https)
	app.wg.Add(1)
	go func(app *App) {
		defer app.wg.Done()
		app.httpsServer.ListenAndServeTLS("", "") // 证书通过GetCertificate回调提供
	}(app)
	return nil
}

func (app *App) ServiceReloadQuicVhostServer(state *State) error {
	app.quicStateLock.Lock()
	defer app.quicStateLock.Unlock()

	// 新配置为关闭vhost server
	if state.Quic == nil || state.Quic.Listen == "" {
		// 老vhost server在运行，就关闭
		if app.quicState != nil && app.quicState.Listen != "" {
			app.quicServer.Close()
			app.quicServer = nil
		}
		app.quicState = state.Quic
		return nil
	}

	// 否则配置为要运行vhost server
	// 判断老的vhost server是否在运行
	// 在运行
	if app.quicState != nil && app.quicState.Listen != "" {
		// 判断监听地址跟新的是否一样
		// 一样就不动老的了
		if app.quicState.Listen == state.Quic.Listen {
			app.quicState = state.Quic
			return nil
		}

		// 不一样，就把老的停掉
		app.quicServer.Close()
		app.quicServer = nil
	}

	// 老的没在运行了
	// 建个新的
	// 新的没跑起来也没办法了，老的没在跑了
	app.quicState = state.Quic
	app.quicServer = app.newQuicServer(state.Quic)
	app.wg.Add(1)
	go func(app *App) {
		defer app.wg.Done()
		app.quicServer.ListenAndServe() // 证书通过GetCertificate回调提供
	}(app)
	return nil
}

func (app *App) ServiceReloadCert(state *State) error {
	app.certStateLock.Lock()
	defer app.certStateLock.Unlock()
	app.certState = state.Cert
	return nil
}
