package config

import (
	"gopkg.in/simversity/gottp.v3/conf"
)

type config struct {
	Gottp conf.GottpSettings
	S3    struct {
		AccessKey string
		SecretKey string
		Bucket    string
		Region    string
	}
	Moire struct {
		DBName               string
		DBAddress            string
		TranslationDirectory string
		Debug                bool
		DBUsername           string
		DBPassword           string
		FFmpeg               string
		SignRequests         bool
		ImageTimeout         int
		StaticPath           string
		Tokens               map[string]string
	}
}

func (self *config) MakeConfig(configPath string) {
	self.Gottp.Listen = "127.0.0.1:8811"
	self.Moire.DBAddress = "127.0.0.1:27017"
	self.Moire.DBName = "gallery"
	self.Moire.FFmpeg = "ffmpeg"
	self.Moire.SignRequests = false
	self.Moire.ImageTimeout = 30
	self.Moire.StaticPath = "https://d3iyv3qavi6f2s.cloudfront.net/statics"

	self.Moire.Tokens = make(map[string]string)
	self.Moire.Tokens[DefaultPublicKey] = DefaultPrivateKey

	self.S3.Region = "eu-west-1"
	self.S3.Bucket = "moire-gallery"

	if configPath != "" {
		conf.MakeConfig(configPath, self)
	}
}

func (self *config) GetGottpConfig() *conf.GottpSettings {
	return &self.Gottp
}

var Settings = config{}
