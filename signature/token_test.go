package signature

import (
	"testing"
	"testing/quick"
)

func seed_token(public_key, private_key string) {
	TOKEN_MAP[public_key] = private_key
}

func TestGetPrivateKey(t *testing.T) {
	f := func(x, y string) bool {
		seed_token(x, y)
		return GetSecretKey(x) == y
	}
	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

func TestDefaultPrivateKey(t *testing.T) {
	key := GetSecretKey("hello")
	if key != defaultPrivateKey {
		t.Error("Key must match the default Secret key")
	}
}
