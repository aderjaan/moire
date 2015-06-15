package signature

var TOKEN_MAP = map[string]string{}

func GetSecretKey(public_key string) string {
	secret, ok := TOKEN_MAP[public_key]
	if !ok {
		return "==GottpMoireToken=="
	}

	return secret
}
