package servercfg

import (
	commoncfg "github.com/polkiloo/go-musthave-metrics-tppl/internal/config/common"
)

const (
	EnvAddressVarName = "ADDRESS"
)

type ServerEnvVars struct {
	Host string // "" если не задано
	Port *int   // nil если не задано
}

func getEnvVars() (ServerEnvVars, error) {
	hp, err := commoncfg.ReadHostPortEnv(EnvAddressVarName)
	if err != nil {
		return ServerEnvVars{}, err
	}
	return ServerEnvVars{Host: hp.Host, Port: hp.Port}, nil
}
