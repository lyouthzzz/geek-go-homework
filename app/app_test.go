package app

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
)

type HTTPServer struct {
}

func (srv *HTTPServer) Endpoint() (string, error) {
	return "http://localhost:8000", nil
}

func (srv *HTTPServer) Start(ctx context.Context) error {
	return errors.New("start error")
}

func (srv *HTTPServer) Stop(ctx context.Context) error {
	return errors.New("stop error")
}

func TestApp(t *testing.T) {
	httpSrv := &HTTPServer{}
	app := NewApp()
	app.AppendServer(httpSrv)

	time.AfterFunc(time.Second, func() {
		app.Stop()
	})
	if err := app.Run(); err != nil {
		t.Fatal(err)
	}

}

func TestErrorGroup(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	eg, cwc := errgroup.WithContext(ctx)

	eg.Go(func() error {
		time.Sleep(5 * time.Second)
		cancel()
		return nil
	})

	eg.Go(func() error {
		return nil
	})

	quitC := make(chan os.Signal, 1)
	signal.Notify(quitC)

	eg.Go(func() error {
		select {
		case <-cwc.Done():
			fmt.Println("context done")
		case <-quitC:
			fmt.Println("quit done")
		}
		fmt.Println("g done")
		return nil
	})

	err := eg.Wait()
	require.NoError(t, err)
}
