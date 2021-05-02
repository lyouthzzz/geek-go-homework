package app

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/sync/errgroup"
)

type Server interface {
	Endpoint() (string, error)
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

type Option func(app *App)

func SignalsOption(sigs []os.Signal) Option {
	return func(app *App) {
		app.sigs = sigs
	}
}

// 参考 kratos app 实现
// https://github.com/go-kratos/kratos/blob/main/app.go
type App struct {
	ctx     context.Context
	cancel  func()
	servers []Server
	sigs    []os.Signal
}

func NewApp(opts ...Option) *App {
	app := &App{
		servers: []Server{},
		sigs: []os.Signal{
			syscall.SIGTERM,
			syscall.SIGQUIT,
			syscall.SIGINT,
		},
	}

	for _, opt := range opts {
		opt(app)
	}

	// 可以由 OptionContext 传入
	app.ctx, app.cancel = context.WithCancel(context.Background())

	return app
}

func (app *App) AppendServer(srv Server) {
	app.servers = append(app.servers, srv)
}

func (app *App) Run() error {
	eg, ctx := errgroup.WithContext(app.ctx)

	for _, srv := range app.servers {
		srv := srv

		eg.Go(func() error {
			// 接受quit信号 -> app.stop() -> app.cancel() 阻塞会取消
			<-ctx.Done()
			return srv.Stop(ctx)
		})

		eg.Go(func() error {
			return srv.Start(ctx)
		})
	}

	quitC := make(chan os.Signal, 1)
	signal.Notify(quitC, app.sigs...)

	eg.Go(func() error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-quitC:
			app.Stop()
		}
		return nil
	})

	return eg.Wait()
}

func (app *App) Stop() {
	// cancel 必须是不为nil的 因为app的停止是要通过cancel去取消的。
	if app.cancel != nil {
		app.cancel()
	}
}
