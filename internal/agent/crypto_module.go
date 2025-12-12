package agent

import (
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/cryptoutil"
	"go.uber.org/fx"
)

// ProvideEncryptor constructs an Encryptor using the configured public key path.
func ProvideEncryptor(cfg AppConfig) (cryptoutil.Encryptor, error) {
	return cryptoutil.NewEncryptorFromPublicKeyFile(cfg.CryptoKeyPath)
}

// ModuleEncryption provides the optional encryptor.
var ModuleEncryption = fx.Module("encryption",
	fx.Provide(
		ProvideEncryptor,
	),
)
