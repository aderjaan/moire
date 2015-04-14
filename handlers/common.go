package handlers

import (
	"github.com/bulletind/moire/config"
	"github.com/bulletind/moire/db"
)

const (
	AudioFile = "audio"
	ImageFile = "image"
	VideoFile = "video"
	PlainFile = "plain"
)

func getConn() *db.MConn {
	return db.GetConn(
		config.Settings.Moire.DBName,
		config.Settings.Moire.DBAddress,
		config.Settings.Moire.DBUsername,
		config.Settings.Moire.DBPassword,
	)
}

func ConcatenateErrors(errs *[]error) string {
	var errString string
	for i := 0; i < len(*errs); i++ {
		errString += (*errs)[i].Error()
		if (len(*errs) - i) > 1 {
			errString += ","
		}
	}
	return errString
}
