package signature

import (
	"github.com/bulletind/moire/config"
)

func GetSecretKey(publicKey string) string {
	// TO DO: replace gcfg setting file with a format, e.g. YAML, that does support maps, so we can do simply this:
	// tokenMap := config.Settings.Moire.Tokens

	tokenMap := make(map[string]string)
	tokenMap[config.Settings.Moire.PublicKey] = config.Settings.Moire.PrivateKey

	secret, ok := tokenMap[publicKey]
	if !ok {
		return config.DefaultPrivateKey
	}

	return secret
}
