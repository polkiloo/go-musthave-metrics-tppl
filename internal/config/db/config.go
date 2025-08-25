package dbcfg

import (
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/db"
	"go.uber.org/fx"
)

func buildDBConfig() (*db.Config, error) {
	env, _ := getEnvVars()
	flags, _ := parseFlags()

	cfg := &db.Config{}

	if env.DSN != "" {
		cfg.DSN = env.DSN
	} else if flags.DSN != "" {
		cfg.DSN = flags.DSN
	}
	return cfg, nil
}

var Module = fx.Module(
	"db-config",
	fx.Provide(buildDBConfig),
)
