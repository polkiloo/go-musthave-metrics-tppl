package dbcfg

import "os"

const EnvDSNVarName = "DATABASE_DSN"

type DBEnvVars struct {
	DSN string
}

func getEnvVars() (DBEnvVars, error) {
	return DBEnvVars{DSN: os.Getenv(EnvDSNVarName)}, nil
}
