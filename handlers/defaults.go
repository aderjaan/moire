package handlers

import (
	"github.com/bulletind/moire/config"
)

const (
	AudioFile = "audio"
	ImageFile = "image"
	VideoFile = "video"
	PlainFile = "plain"
)

const DefaultCollection = "assets"
const UploadPrefix = "original_file"

func thumbnailUrls(fileType string) string {
	switch fileType {
	case AudioFile:
		return config.Settings.Moire.StaticPath + "/placeholders/audio.png"
	case VideoFile:
		return config.Settings.Moire.StaticPath + "/placeholders/video.png"
	case ImageFile:
		return config.Settings.Moire.StaticPath + "/placeholders/thumbnail.png"
	default:
		return config.Settings.Moire.StaticPath + "/placeholders/attachment.png"
	}
}

const missingUrl = "https://s3-eu-west-1.amazonaws.com/sc-gallery/missing.png"

const thumbUnavailable = "Ouch! This thumbnail is no longer available."
const uploadInProgress = "This content is still being uploaded. We appreciate your impatience"
const contentUnavailable = "Ouch! This content is no longer available."
