package server

import (
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/cryptoutil"
	"go.uber.org/fx"
)

// ProvideDecryptor creates a decryptor using the configured private key path.
func ProvideDecryptor(cfg AppConfig) (cryptoutil.Decryptor, error) {
	return cryptoutil.NewDecryptorFromPrivateKeyFile(cfg.CryptoKeyPath)
}

// ModuleCrypto wires the crypto dependencies for the server.
var ModuleCrypto = fx.Module(
	"server-crypto",
	fx.Provide(
		ProvideDecryptor,
	),
)
