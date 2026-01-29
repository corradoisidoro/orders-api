package application_test

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/corradoisidoro/orders-api/internal/application"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type fakeServer struct {
	listenErr   error
	shutdownErr error
}

func (f *fakeServer) ListenAndServe() error {
	return f.listenErr
}

func (f *fakeServer) Shutdown(ctx context.Context) error {
	return f.shutdownErr
}

func TestAppStart_ServerError(t *testing.T) {
	cfg := application.Config{
		ServerPort:  9999,
		DatabaseDSN: "dsn",
	}

	var db *gorm.DB
	app := application.New(cfg, db)

	fake := &fakeServer{
		listenErr: errors.New("boom"),
	}

	app.ServerFactory = func(addr string, h http.Handler) application.HTTPServer {
		return fake
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := app.Start(ctx)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "server error")
	assert.Contains(t, err.Error(), "boom")
}

func TestAppStart_GracefulShutdown(t *testing.T) {
	cfg := application.Config{
		ServerPort:  9998,
		DatabaseDSN: "dsn",
	}

	var db *gorm.DB
	app := application.New(cfg, db)

	// http.ErrServerClosed is treated as a normal shutdown by App.Start.
	fake := &fakeServer{
		listenErr: http.ErrServerClosed,
	}

	app.ServerFactory = func(addr string, h http.Handler) application.HTTPServer {
		return fake
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	err := app.Start(ctx)

	require.NoError(t, err)
}
