package dbcfg

import (
	"os"

	commoncfg "github.com/polkiloo/go-musthave-metrics-tppl/internal/config/common"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/db"
	"go.uber.org/fx"
)

func buildDBConfig() (*db.Config, error) {
	env, _ := getEnvVars()
	flags, _ := parseFlags()

	configPath := os.Getenv("CONFIG")
	if configPath == "" {
		configPath = flags.ConfigPath
	}

	var fileCfg dbFileConfig
	if err := commoncfg.LoadConfigFile(configPath, &fileCfg); err != nil {
		return nil, err
	}

	dsn := ""
	if fileCfg.DatabaseDSN != nil {
		dsn = *fileCfg.DatabaseDSN
	}

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
