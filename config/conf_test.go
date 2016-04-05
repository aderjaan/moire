package config

import (
	"gopkg.in/simversity/gottp.v3"
	"testing"
)

func TestConfig(t *testing.T) {
	gottp.MakeConfig(&Settings)

	if Settings.Moire.DbUrl != "mongodb://localhost/gallery" {
		t.Error("Improper Configuration: " + Settings.Moire.DbUrl)
	}
}
