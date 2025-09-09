package dbcfg

import (
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/db"
	"go.uber.org/fx"
)

func buildDBConfig() (*db.Config, error) {
	env, _ := getEnvVars()
	flags, _ := parseFlags()

	dsn := ""
	if env.DSN != "" {
		dsn = env.DSN
	} else if flags.DSN != "" {
		dsn = flags.DSN
	}
	return &db.Config{DSN: dsn}, nil
}

var Module = fx.Module(
	"db-config",
	fx.Provide(buildDBConfig),
)
