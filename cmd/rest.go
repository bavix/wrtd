package cmd

import (
	"context"
	"errors"
	"net"
	"net/http"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/bavix/wrtd/internal/app"
)

func NewRestServer(
	lc fx.Lifecycle,
	mux *http.ServeMux,
	c *app.Checker,
	log *zap.Logger,
) *http.Server {
	srv := &http.Server{Addr: ":3333", Handler: mux} //nolint:gosec

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			ln, err := net.Listen("tcp", srv.Addr)
			if err != nil {
				return err
			}

			log.Info("Starting HTTP server", zap.String("addr", srv.Addr))
			go func() {
				err := srv.Serve(ln)
				if !errors.Is(err, http.ErrServerClosed) {
					log.Error("Server error", zap.Error(err))
				}
			}()

			log.Info("Starting checker")
			go c.Run(context.WithoutCancel(ctx))

			return nil
		},
		OnStop: func(ctx context.Context) error {
			return srv.Shutdown(ctx)
		},
	})

	return srv
}
