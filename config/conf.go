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
		DBName                 string
		DBAddress              string
		TranslationDirectory   string
		Debug                  bool
		DBUsername             string
		DBPassword             string
		FFmpeg                 string
		SignRequests           bool
		ImageTimeout           int
		StaticPath             string
		PublicKey              string
		PrivateKey             string
		SentryDSN              string
		UploadUrlExpiry        int64
		GetUrlExpiry           int64
		RedirectUrlCacheExpiry int64
	}
}

func (self *config) MakeConfig(configPath string) {
	self.Gottp.Listen = "127.0.0.1:8811"
	self.Moire.DBAddress = "127.0.0.1:27017"
	self.Moire.DBName = "gallery"
	self.Moire.Debug = false
	self.Moire.FFmpeg = "ffmpeg"
	self.Moire.SignRequests = false
	self.Moire.ImageTimeout = 15
	self.Moire.StaticPath = "https://d3iyv3qavi6f2s.cloudfront.net/statics"

	self.Moire.PublicKey = DefaultPublicKey
	self.Moire.PrivateKey = DefaultPrivateKey
	self.Moire.UploadUrlExpiry = 7200      // 5 days (60 * 24 * 5)
	self.Moire.GetUrlExpiry = 60           // 1 hour
	self.Moire.RedirectUrlCacheExpiry = 45 // 45 minutes, must be lower than GetUrlExpiry

	self.S3.Region = "eu-west-1"
	self.S3.Bucket = "moire-gallery"

	if configPath != "" {
		conf.MakeConfig(configPath, self)
	}
}

func (self *config) GetGottpConfig() *conf.GottpSettings {
	return &self.Gottp
}

var Settings config
