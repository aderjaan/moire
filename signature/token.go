package signature

import (
	"github.com/bulletind/moire/config"
)

func GetSecretKey(publicKey string) string {
	tokenMap := config.Settings.Moire.Tokens
	secret, ok := tokenMap[publicKey]
	if !ok {
		return config.DefaultPrivateKey
	}

	return secret
}
