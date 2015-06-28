package signature

var TOKEN_MAP = map[string]string{}

const defaultPrivateKey = "==GottpMoireToken=="

func GetSecretKey(public_key string) string {
	secret, ok := TOKEN_MAP[public_key]
	if !ok {
		return defaultPrivateKey
	}

	return secret
}
