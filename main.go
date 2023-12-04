package main

import (
	"net/http"

	"go.uber.org/config"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"

	"github.com/bavix/wrtd/cmd"
	"github.com/bavix/wrtd/internal/app"
)

func NewConfig() (*config.YAML, error) {
	yaml, err := config.NewYAML(
		config.File("/etc/wrtd/config.yaml"),
	)
	if err != nil {
		return nil, err
	}

	return yaml, nil
}

func main() {
	fx.New(
		fx.WithLogger(func(log *zap.Logger) fxevent.Logger {
			return &fxevent.ZapLogger{Logger: log}
		}),
		fx.Provide(
			NewConfig,
			cmd.NewRestServer,
			app.NewServeMux,
			app.NewCheckList,
			app.NewChecker,
			zap.NewProduction,
		),
		fx.Invoke(func(*http.Server) {}),
		fx.NopLogger,
	).Run()
}
