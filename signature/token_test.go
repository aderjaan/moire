package signature

import (
	"github.com/bulletind/moire/config"
	"gopkg.in/simversity/gottp.v3"
	"testing"
	"testing/quick"
)

func seed_token(publicKey, privateKey string) {
	config.Settings.Moire.Tokens[publicKey] = privateKey
}

func init() {
	gottp.MakeConfig(&config.Settings)
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
	if key != config.DefaultPrivateKey {
		t.Error("Key must match the default Secret key")
	}
}
