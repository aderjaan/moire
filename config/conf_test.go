package config

import (
	"gopkg.in/simversity/gottp.v3"
	"testing"
)

func TestConfig(t *testing.T) {
	gottp.MakeConfig(&Settings)

	if Settings.Moire.DBAddress != "127.0.0.1:27017" {
		t.Error("Improper Configuration")
	}
}
